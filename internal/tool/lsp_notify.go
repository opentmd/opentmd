package tool

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
)

func (r *Runtime) notifyLSPAfterTool(ctx context.Context, tool, argsJSON string) {
	if r.lsp == nil {
		return
	}
	path := extractPathArg(tool, argsJSON)
	if path == "" {
		return
	}
	abs := path
	if !filepath.IsAbs(abs) {
		abs = filepath.Join(r.WorkDir(), path)
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		return
	}
	r.lsp.NotifyFileChanged(ctx, abs, string(data))
}

func extractPathArg(tool, argsJSON string) string {
	type pathFields struct {
		Path     string `json:"path"`
		FilePath string `json:"file_path"`
	}
	var f pathFields
	if err := json.Unmarshal([]byte(argsJSON), &f); err != nil {
		return ""
	}
	switch tool {
	case "write_file", "edit_file", "search_replace":
		if f.Path != "" {
			return f.Path
		}
		return f.FilePath
	default:
		return ""
	}
}
