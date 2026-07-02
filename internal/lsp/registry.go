package lsp

import (
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/opentmd/opentmd-cli/internal/config"
)

type ServerConfig struct {
	Command     string
	Args        []string
	RootMarkers []string
}

type Registry struct {
	servers map[string]ServerConfig
}

func NewRegistry(autoDetect bool, user map[string]config.LSPServerConfig) *Registry {
	r := &Registry{servers: map[string]ServerConfig{}}
	if autoDetect {
		r.mergeDefaults()
	}
	for ext, sc := range user {
		if sc.Command == "" {
			continue
		}
		r.servers[normalizeExt(ext)] = ServerConfig{
			Command:     sc.Command,
			Args:        sc.Args,
			RootMarkers: sc.RootMarkers,
		}
	}
	return r
}

func normalizeExt(ext string) string {
	ext = strings.ToLower(strings.TrimPrefix(ext, "."))
	return ext
}

func (r *Registry) mergeDefaults() {
	defaults := map[string]ServerConfig{
		"rs":  {Command: "rust-analyzer", Args: nil, RootMarkers: []string{"Cargo.toml"}},
		"ts":  {Command: "typescript-language-server", Args: []string{"--stdio"}, RootMarkers: []string{"tsconfig.json", "package.json"}},
		"tsx": {Command: "typescript-language-server", Args: []string{"--stdio"}, RootMarkers: []string{"tsconfig.json"}},
		"js":  {Command: "typescript-language-server", Args: []string{"--stdio"}, RootMarkers: []string{"package.json"}},
		"py":  {Command: "pylsp", Args: nil, RootMarkers: []string{"pyproject.toml", "setup.py"}},
		"go":  {Command: "gopls", Args: []string{"serve"}, RootMarkers: []string{"go.mod"}},
		"java": {Command: "jdtls", Args: nil, RootMarkers: []string{"pom.xml", "build.gradle"}},
	}
	for k, v := range defaults {
		r.servers[k] = v
	}
}

func (r *Registry) Get(ext string) (ServerConfig, bool) {
	cfg, ok := r.servers[normalizeExt(ext)]
	return cfg, ok
}

func (r *Registry) Extensions() []string {
	exts := make([]string, 0, len(r.servers))
	for ext := range r.servers {
		exts = append(exts, ext)
	}
	sort.Strings(exts)
	return exts
}

func (r *Registry) ForFile(path string) (ServerConfig, bool) {
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
	return r.Get(ext)
}

func commandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func languageID(ext string) string {
	switch ext {
	case "rs":
		return "rust"
	case "ts":
		return "typescript"
	case "tsx":
		return "typescriptreact"
	case "js":
		return "javascript"
	case "jsx":
		return "javascriptreact"
	case "py":
		return "python"
	case "go":
		return "go"
	case "java":
		return "java"
	case "c":
		return "c"
	case "cpp", "cc", "cxx":
		return "cpp"
	case "cs":
		return "csharp"
	default:
		return ext
	}
}
