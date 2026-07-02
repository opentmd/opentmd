package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/opentmd/opentmd-cli/internal/config"
)

func TestEnsureCreatesLayout(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	if err := config.Ensure(); err != nil {
		t.Fatalf("ensure: %v", err)
	}

	root := filepath.Join(dir, ".opentmd")
	for _, name := range []string{"config.toml", "sessions", "plugins", "skills", "mcp.json", "hooks.json"} {
		path := filepath.Join(root, name)
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected %s to exist: %v", path, err)
		}
	}
}

func TestLoadPreservesDefaultProviders(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	opentmdDir := filepath.Join(dir, ".opentmd")
	if err := os.MkdirAll(opentmdDir, 0o755); err != nil {
		t.Fatal(err)
	}
	cfgPath := filepath.Join(opentmdDir, "config.toml")
	if err := os.WriteFile(cfgPath, []byte(`
[default]
provider = "deepseek"
model = "deepseek-chat"

[providers]
[session]
persist = true
[tui]
theme = "dark"
stream = true
`), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	for _, name := range []string{"deepseek", "openai", "glm", "qwen", "ollama"} {
		if _, ok := cfg.Providers[name]; !ok {
			t.Fatalf("expected provider %q", name)
		}
	}
}

func TestNeedsWizard(t *testing.T) {
	cfg := config.DefaultConfig()
	if !cfg.NeedsWizard() {
		t.Fatal("fresh config should need wizard")
	}
	cfg.Providers["deepseek"] = config.ProviderConfig{
		BaseURL: "https://api.deepseek.com",
		APIKey:  "sk-test",
	}
	cfg.Setup.Completed = true
	if cfg.NeedsWizard() {
		t.Fatal("configured config should not need wizard")
	}
}

func TestLSPDefaults(t *testing.T) {
	cfg := config.DefaultConfig()
	if !cfg.LSP.Enabled {
		t.Fatal("LSP should be enabled by default")
	}
	if !cfg.LSP.AutoDetect {
		t.Fatal("LSP auto_detect should be true by default")
	}
	if cfg.LSP.DiagnosticsSettleDelayMs != 300 {
		t.Fatalf("settle delay: got %d", cfg.LSP.DiagnosticsSettleDelayMs)
	}
	if cfg.LSP.WarmupFileLimit != 40 {
		t.Fatalf("warmup limit: got %d", cfg.LSP.WarmupFileLimit)
	}
}
