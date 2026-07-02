package graph

import (
	"encoding/gob"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/opentmd/opentmd/internal/pathutil"
	"github.com/opentmd/opentmd/internal/project"
)

type Symbol struct {
	Name string
	File string
	Line int
	Kind string
}

type Edge struct {
	From string
	To   string
	Kind string
}

type Graph struct {
	Symbols []Symbol
	Edges   []Edge
}

type Index struct {
	mu      sync.RWMutex
	graph   Graph
	workDir string
}

func NewIndex(workDir string) *Index {
	return &Index{workDir: workDir}
}

func (idx *Index) LoadOrBuild() error {
	path := project.GraphCachePath(idx.workDir)
	if f, err := os.Open(path); err == nil {
		var g Graph
		if err := gob.NewDecoder(f).Decode(&g); err == nil && len(g.Symbols) > 0 {
			_ = f.Close()
			idx.mu.Lock()
			idx.graph = g
			idx.mu.Unlock()
			return nil
		}
		_ = f.Close()
	}
	return idx.Build()
}

func (idx *Index) Build() error {
	fset := token.NewFileSet()
	var symbols []Symbol
	var edges []Edge

	_ = filepath.WalkDir(idx.workDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			if d != nil && d.IsDir() && path != idx.workDir && pathutil.ShouldSkipDir(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		rel, _ := filepath.Rel(idx.workDir, path)
		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return nil
		}
		calls := map[string]string{}
		ast.Inspect(f, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.FuncDecl:
				if x.Name != nil {
					id := rel + ":" + x.Name.Name
					pos := fset.Position(x.Pos())
					symbols = append(symbols, Symbol{Name: x.Name.Name, File: rel, Line: pos.Line, Kind: "func"})
					calls[id] = x.Name.Name
				}
			case *ast.CallExpr:
				if sel, ok := x.Fun.(*ast.SelectorExpr); ok {
					if id, ok := sel.X.(*ast.Ident); ok {
						caller := currentFunc(fset, f, x)
						if caller != "" {
							edges = append(edges, Edge{From: caller, To: id.Name + "." + sel.Sel.Name, Kind: "call"})
						}
					}
				}
			}
			return true
		})
		_ = calls
		return nil
	})

	idx.mu.Lock()
	idx.graph = Graph{Symbols: symbols, Edges: edges}
	idx.mu.Unlock()

	_ = os.MkdirAll(project.Dir(idx.workDir), 0o755)
	return saveGob(project.GraphCachePath(idx.workDir), idx.graph)
}

func currentFunc(fset *token.FileSet, f *ast.File, n ast.Node) string {
	var fn string
	var best ast.Node
	ast.Inspect(f, func(node ast.Node) bool {
		if fd, ok := node.(*ast.FuncDecl); ok {
			if fd.Pos() <= n.Pos() && n.Pos() <= fd.End() {
				if best == nil || fd.Pos() > best.Pos() {
					best = fd
					fn = fd.Name.Name
				}
			}
		}
		return true
	})
	return fn
}

func saveGob(path string, g Graph) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return gob.NewEncoder(f).Encode(g)
}

func (idx *Index) TraceCallers(symbol string) string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	var out []string
	for _, e := range idx.graph.Edges {
		if strings.Contains(e.To, symbol) || e.To == symbol {
			out = append(out, e.From+" -> "+e.To)
		}
	}
	if len(out) == 0 {
		return "no callers found for " + symbol
	}
	return strings.Join(out, "\n")
}

func (idx *Index) TraceCallees(symbol string, depth int) string {
	if depth <= 0 {
		depth = 3
	}
	if depth > 5 {
		depth = 5
	}

	idx.mu.RLock()
	defer idx.mu.RUnlock()

	var roots []string
	seenRoot := map[string]bool{}
	for _, s := range idx.graph.Symbols {
		if s.Name != symbol {
			continue
		}
		key := s.File + ":" + s.Name
		if seenRoot[key] {
			continue
		}
		seenRoot[key] = true
		roots = append(roots, s.Name)
	}
	if len(roots) == 0 {
		for _, e := range idx.graph.Edges {
			if e.From == symbol {
				roots = []string{symbol}
				break
			}
		}
	}
	if len(roots) == 0 {
		return "symbol not found in code graph: " + symbol
	}

	var sb strings.Builder
	for _, root := range roots {
		sb.WriteString("Callees of ")
		sb.WriteString(root)
		sb.WriteString(":\n")
		type node struct {
			name  string
			depth int
		}
		visited := map[string]bool{}
		queue := []node{{name: root, depth: 0}}
		found := false
		for len(queue) > 0 {
			cur := queue[0]
			queue = queue[1:]
			if cur.depth >= depth {
				continue
			}
			for _, e := range idx.graph.Edges {
				if e.From != cur.name {
					continue
				}
				if visited[e.To] {
					continue
				}
				visited[e.To] = true
				found = true
				indent := strings.Repeat("  ", cur.depth+1)
				fmt.Fprintf(&sb, "%s[depth %d] %s\n", indent, cur.depth+1, e.To)
				queue = append(queue, node{name: e.To, depth: cur.depth + 1})
			}
		}
		if !found {
			sb.WriteString("  (no callees found)\n")
		}
		sb.WriteByte('\n')
	}
	return strings.TrimSpace(sb.String())
}

func (idx *Index) FileDependencies(file string) string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	var out []string
	for _, s := range idx.graph.Symbols {
		if s.File == file || strings.HasSuffix(s.File, file) {
			out = append(out, fmt.Sprintf("%s %s @ %s:%d", s.Kind, s.Name, s.File, s.Line))
		}
	}
	if len(out) == 0 {
		return "no symbols in " + file
	}
	return strings.Join(out, "\n")
}
