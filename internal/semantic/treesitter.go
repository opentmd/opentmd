package semantic

import (
	"fmt"
	"sort"

	"github.com/odvcencio/gotreesitter"
	"github.com/odvcencio/gotreesitter/grammars"
)

func listSymbolsTreeSitter(path string, source []byte) ([]Symbol, error) {
	entry := grammars.DetectLanguage(path)
	if entry == nil {
		return nil, fmt.Errorf("tree-sitter: unsupported language for %s", path)
	}

	querySrc, err := symbolsQuery(entry.Name)
	if err != nil {
		return nil, err
	}

	lang := entry.Language()
	parser := gotreesitter.NewParser(lang)
	tree, err := parser.Parse(source)
	if err != nil {
		return nil, fmt.Errorf("tree-sitter parse: %w", err)
	}
	defer tree.Release()

	root := tree.RootNode()
	if root == nil {
		return nil, fmt.Errorf("tree-sitter: empty parse tree")
	}

	q, err := gotreesitter.NewQuery(querySrc, lang)
	if err != nil {
		return nil, fmt.Errorf("tree-sitter query: %w", err)
	}

	type key struct {
		name string
		line int
	}
	seen := map[key]struct{}{}
	var symbols []Symbol

	for _, match := range q.Execute(tree) {
		var defNode *gotreesitter.Node
		name := ""
		for _, cap := range match.Captures {
			switch cap.Name {
			case "definition":
				defNode = cap.Node
			case "name":
				if cap.Node != nil {
					name = cap.Text(source)
				}
			}
		}
		if defNode == nil || name == "" {
			continue
		}

		startLine := int(defNode.StartPoint().Row) + 1
		endLine := int(defNode.EndPoint().Row) + 1
		if endLine < startLine {
			endLine = startLine
		}

		k := key{name: name, line: startLine}
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}

		symbols = append(symbols, Symbol{
			Name:      name,
			Kind:      defNode.Type(lang),
			StartLine: startLine,
			EndLine:   endLine,
		})
	}

	sort.Slice(symbols, func(i, j int) bool {
		if symbols[i].StartLine == symbols[j].StartLine {
			return symbols[i].Name < symbols[j].Name
		}
		return symbols[i].StartLine < symbols[j].StartLine
	})

	return symbols, nil
}
