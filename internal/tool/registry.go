package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/opentmd/opentmd/internal/fileio"
	"github.com/opentmd/opentmd/internal/glob"
	"github.com/opentmd/opentmd/internal/grep"
	"github.com/opentmd/opentmd/internal/llm"
	"github.com/opentmd/opentmd/internal/pathutil"
	"github.com/opentmd/opentmd/internal/shell"
)

type Handler func(ctx context.Context, args json.RawMessage) (string, error)

type Registry struct {
	workDir     string
	tools       map[string]toolEntry
	beforeWrite func(path string) error
}

func (r *Registry) SetBeforeWrite(fn func(path string) error) {
	r.beforeWrite = fn
}

type toolEntry struct {
	def    llm.ToolDefinition
	handle Handler
}

func NewRegistry(workDir string) *Registry {
	if workDir == "" {
		workDir, _ = os.Getwd()
	}
	r := &Registry{workDir: workDir, tools: map[string]toolEntry{}}
	r.registerDefaults()
	return r
}

func (r *Registry) WorkDir() string {
	return r.workDir
}

func (r *Registry) SetWorkDir(dir string) {
	r.workDir = dir
}

func (r *Registry) registerDefaults() {
	r.Register(llm.ToolDefinition{
		Name: "read_file",
		Description: "Read a local file. Large files are truncated. Use grep to search content, glob to find files.",
		Parameters: objParam(map[string]any{
			"path": strParam("File path relative to project root"),
		}, "path"),
	}, r.readFile)

	r.Register(llm.ToolDefinition{
		Name: "write_file",
		Description: "Create or overwrite a file with the given content.",
		Parameters: objParam(map[string]any{
			"path":    strParam("File path"),
			"content": strParam("Full file content"),
		}, "path", "content"),
	}, r.writeFile)

	r.Register(llm.ToolDefinition{
		Name: "edit_file",
		Description: "Replace the first occurrence of old_string with new_string in a single file.",
		Parameters: objParam(map[string]any{
			"path":       strParam("File path"),
			"old_string": strParam("Exact text to find"),
			"new_string": strParam("Replacement text"),
		}, "path", "old_string", "new_string"),
	}, r.editFile)

	r.Register(llm.ToolDefinition{
		Name: "grep",
		Description: "Search file contents by regex pattern. Returns matching lines with context.",
		Parameters: objParam(map[string]any{
			"pattern":     strParam("Regex pattern (case-insensitive unless uppercase used)"),
			"path":        strParam("File or directory (default: project root)"),
			"max_results": intParam("Max results (default 50)"),
		}, "pattern"),
	}, r.grepTool)

	r.Register(llm.ToolDefinition{
		Name: "glob",
		Description: "Find files by name pattern, e.g. **/*.go, src/**/*.ts",
		Parameters: objParam(map[string]any{
			"pattern": strParam("Glob pattern"),
			"path":    strParam("Base directory (default: project root)"),
		}, "pattern"),
	}, r.globTool)

	r.Register(llm.ToolDefinition{
		Name: "list_directory",
		Description: "List files and directories in a path.",
		Parameters: objParam(map[string]any{
			"path": strParam("Directory path (default: project root)"),
		}),
	}, r.listDir)

	r.Register(llm.ToolDefinition{
		Name: "bash",
		Description: "Run a shell command in the project directory. Use for tests, builds, git, etc.",
		Parameters: objParam(map[string]any{
			"command": strParam("Shell command"),
		}, "command"),
	}, r.bash)

	r.Register(llm.ToolDefinition{
		Name: "search_replace",
		Description: "Replace all occurrences of search with replace across matching files.",
		Parameters: objParam(map[string]any{
			"search":  strParam("Text to find"),
			"replace": strParam("Replacement"),
			"glob":    strParam("File pattern e.g. *.go (optional)"),
			"path":    strParam("Directory (default: project root)"),
		}, "search", "replace"),
	}, r.searchReplace)

	// backward compat aliases (execute only, not in Definitions)
	r.tools["run_shell"] = r.tools["bash"]
	r.tools["list_dir"] = r.tools["list_directory"]
}

func objParam(props map[string]any, required ...string) map[string]any {
	m := map[string]any{"type": "object", "properties": props}
	if len(required) > 0 {
		m["required"] = required
	}
	return m
}

func strParam(desc string) map[string]any {
	return map[string]any{"type": "string", "description": desc}
}

func intParam(desc string) map[string]any {
	return map[string]any{"type": "integer", "description": desc}
}

func (r *Registry) Register(def llm.ToolDefinition, handle Handler) {
	r.tools[def.Name] = toolEntry{def: def, handle: handle}
}

func (r *Registry) Definitions() []llm.ToolDefinition {
	seen := map[string]bool{}
	var out []llm.ToolDefinition
	for name, t := range r.tools {
		if seen[name] || name == "run_shell" || name == "list_dir" {
			continue
		}
		seen[name] = true
		out = append(out, t.def)
	}
	return out
}

func (r *Registry) Execute(ctx context.Context, call llm.ToolCall) (string, error) {
	entry, ok := r.tools[call.Name]
	if !ok {
		return "", fmt.Errorf("unknown tool %q", call.Name)
	}
	return entry.handle(ctx, json.RawMessage(call.Arguments))
}

func (r *Registry) resolve(p string) (string, error) {
	return pathutil.Resolve(r.workDir, p)
}

func parseArgs[T any](args json.RawMessage, tool string) (T, error) {
	var v T
	if err := json.Unmarshal(args, &v); err != nil {
		var zero T
		return zero, fmt.Errorf("%s: invalid arguments: %w (check required fields)", tool, err)
	}
	return v, nil
}

type pathArg struct {
	Path string `json:"path"`
}

func (r *Registry) readFile(ctx context.Context, args json.RawMessage) (string, error) {
	a, err := parseArgs[pathArg](args, "read_file")
	if err != nil {
		return "", err
	}
	path, err := r.resolve(a.Path)
	if err != nil {
		return "", err
	}
	return fileio.Read(path, fileio.DefaultMaxBytes)
}

type writeArgs struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

func (r *Registry) writeFile(ctx context.Context, args json.RawMessage) (string, error) {
	a, err := parseArgs[writeArgs](args, "write_file")
	if err != nil {
		return "", err
	}
	path, err := r.resolve(a.Path)
	if err != nil {
		return "", err
	}
	if r.beforeWrite != nil {
		if err := r.beforeWrite(path); err != nil {
			return "", fmt.Errorf("backup: %w", err)
		}
	}
	if err := fileio.Write(path, a.Content); err != nil {
		return "", err
	}
	rel, _ := filepath.Rel(r.workDir, path)
	return fmt.Sprintf("wrote %s (%d bytes)", rel, len(a.Content)), nil
}

type editArgs struct {
	Path      string `json:"path"`
	OldString string `json:"old_string"`
	NewString string `json:"new_string"`
}

func (r *Registry) editFile(ctx context.Context, args json.RawMessage) (string, error) {
	a, err := parseArgs[editArgs](args, "edit_file")
	if err != nil {
		return "", err
	}
	path, err := r.resolve(a.Path)
	if err != nil {
		return "", err
	}
	if r.beforeWrite != nil {
		if err := r.beforeWrite(path); err != nil {
			return "", fmt.Errorf("backup: %w", err)
		}
	}
	preview, err := fileio.Edit(path, a.OldString, a.NewString)
	if err != nil {
		return "", err
	}
	rel, _ := filepath.Rel(r.workDir, path)
	return fmt.Sprintf("edited %s\n%s", rel, preview), nil
}

type grepArgs struct {
	Pattern    string `json:"pattern"`
	Path       string `json:"path"`
	MaxResults int    `json:"max_results"`
}

func (r *Registry) grepTool(ctx context.Context, args json.RawMessage) (string, error) {
	a, err := parseArgs[grepArgs](args, "grep")
	if err != nil {
		return "", err
	}
	root := r.workDir
	if a.Path != "" {
		root, err = r.resolve(a.Path)
		if err != nil {
			return "", err
		}
	}
	return grep.Search(grep.Options{Pattern: a.Pattern, Root: root, MaxResults: a.MaxResults})
}

type globArgs struct {
	Pattern string `json:"pattern"`
	Path    string `json:"path"`
}

func (r *Registry) globTool(ctx context.Context, args json.RawMessage) (string, error) {
	a, err := parseArgs[globArgs](args, "glob")
	if err != nil {
		return "", err
	}
	root := r.workDir
	if a.Path != "" {
		root, err = r.resolve(a.Path)
		if err != nil {
			return "", err
		}
	}
	return glob.Find(glob.Options{Pattern: a.Pattern, Root: root})
}

type listDirArgs struct {
	Path string `json:"path"`
}

func (r *Registry) listDir(ctx context.Context, args json.RawMessage) (string, error) {
	a, err := parseArgs[listDirArgs](args, "list_directory")
	if err != nil {
		return "", err
	}
	dir := r.workDir
	if a.Path != "" {
		dir, err = r.resolve(a.Path)
		if err != nil {
			return "", err
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("list dir: %w", err)
	}
	var sb strings.Builder
	for _, e := range entries {
		if e.IsDir() {
			sb.WriteString("[dir]  ")
		} else {
			sb.WriteString("[file] ")
		}
		sb.WriteString(e.Name())
		sb.WriteString("\n")
	}
	return sb.String(), nil
}

type bashArgs struct {
	Command string `json:"command"`
}

func (r *Registry) bash(ctx context.Context, args json.RawMessage) (string, error) {
	a, err := parseArgs[bashArgs](args, "bash")
	if err != nil {
		return "", err
	}
	res, err := shell.Run(ctx, a.Command, r.workDir, shell.DefaultTimeout, shell.DefaultMaxOutput)
	if err != nil {
		return "", err
	}
	return shell.FormatResult(res), nil
}

func (r *Registry) searchReplace(ctx context.Context, args json.RawMessage) (string, error) {
	a, err := parseArgs[searchReplaceArgs](args, "search_replace")
	if err != nil {
		return "", err
	}
	root := r.workDir
	if a.Path != "" {
		root, err = r.resolve(a.Path)
		if err != nil {
			return "", err
		}
	}
	pattern := a.Glob
	if pattern == "" {
		pattern = "**/*"
	}
	listing, err := glob.Find(glob.Options{Pattern: pattern, Root: root, Limit: 500})
	if err != nil {
		return "", err
	}
	if listing == "no files matched" {
		return "no files matched", nil
	}
	var changed []string
	for _, rel := range strings.Split(listing, "\n") {
		if rel == "" || strings.HasPrefix(rel, "(") {
			continue
		}
		path, err := r.resolve(rel)
		if err != nil {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		content := string(data)
		if !strings.Contains(content, a.Search) {
			continue
		}
		updated := strings.ReplaceAll(content, a.Search, a.Replace)
		if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
			return "", err
		}
		changed = append(changed, rel)
	}
	if len(changed) == 0 {
		return "no occurrences found", nil
	}
	return fmt.Sprintf("updated %d file(s):\n%s", len(changed), strings.Join(changed, "\n")), nil
}

type searchReplaceArgs struct {
	Search  string `json:"search"`
	Replace string `json:"replace"`
	Glob    string `json:"glob"`
	Path    string `json:"path"`
}
