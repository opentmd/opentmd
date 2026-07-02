package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

const (
	DefaultProvider = "deepseek"
	DefaultModel    = "deepseek-chat"
	DefaultTheme    = "dark"
)

type ProviderConfig struct {
	BaseURL string `toml:"base_url"`
	APIKey  string `toml:"api_key"`
}

type Config struct {
	Default struct {
		Provider string `toml:"provider"`
		Model    string `toml:"model"`
	} `toml:"default"`

	Providers map[string]ProviderConfig `toml:"providers"`

	Session struct {
		Persist bool `toml:"persist"`
	} `toml:"session"`

	TUI struct {
		Theme  string `toml:"theme"`
		Stream bool   `toml:"stream"`
	} `toml:"tui"`

	Agent struct {
		MaxIterations int `toml:"max_iterations"`
	} `toml:"agent"`

	Setup struct {
		Completed bool `toml:"completed"`
	} `toml:"setup"`

	LSP LSPConfig `toml:"lsp"`

	DefaultWorkdir string `toml:"default_workdir"`
}

func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	return filepath.Join(home, ".opentmd"), nil
}

// LayoutDirs lists subdirectories created under ~/.opentmd on startup.
var LayoutDirs = []string{"sessions", "plugins", "skills"}

func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.toml"), nil
}

func SessionsDir() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "sessions"), nil
}

func PluginsDir() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "plugins"), nil
}

// Ensure creates ~/.opentmd, required subdirectories, and default config.toml.
func Ensure() error {
	root, err := Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return fmt.Errorf("create opentmd dir: %w", err)
	}
	for _, name := range LayoutDirs {
		sub := filepath.Join(root, name)
		if err := os.MkdirAll(sub, 0o755); err != nil {
			return fmt.Errorf("create %q: %w", sub, err)
		}
	}
	cfgPath, err := Path()
	if err != nil {
		return err
	}
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		if err := Save(DefaultConfig()); err != nil {
			return err
		}
	}
	if err := ensureMCPConfig(root); err != nil {
		return err
	}
	if err := ensureHooksConfig(root); err != nil {
		return err
	}
	if err := ensureExampleSkill(root); err != nil {
		return err
	}
	return nil
}

func ensureMCPConfig(root string) error {
	path := filepath.Join(root, "mcp.json")
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	example := `{
  "mcpServers": {
    "example": {
      "command": "echo",
      "args": ["MCP server placeholder — replace with real command"],
      "disabled": true
    }
  }
}
`
	return os.WriteFile(path, []byte(example), 0o600)
}

func ensureHooksConfig(root string) error {
	path := filepath.Join(root, "hooks.json")
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	example := `{
  "hooks": {
    "pre_bash": {
      "event": "PreToolUse",
      "matcher": "bash",
      "command": "echo '{\"action\":\"allow\"}'",
      "disabled": true,
      "timeout_ms": 5000
    }
  }
}
`
	return os.WriteFile(path, []byte(example), 0o644)
}

func ensureExampleSkill(root string) error {
	dir := filepath.Join(root, "skills", "example")
	path := filepath.Join(dir, "SKILL.md")
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	example := `---
name: example
description: "示例 Skill — 展示 SKILL.md 格式"
---

# Example Skill

用户参数: $ARGUMENTS

请按上述参数完成任务。
`
	return os.WriteFile(path, []byte(example), 0o644)
}

func DefaultConfig() *Config {
	cfg := &Config{}
	cfg.Default.Provider = DefaultProvider
	cfg.Default.Model = DefaultModel
	cfg.Providers = defaultProvidersMap()
	cfg.Session.Persist = true
	cfg.TUI.Theme = DefaultTheme
	cfg.TUI.Stream = true
	cfg.Agent.MaxIterations = 10
	cfg.Setup.Completed = false
	cfg.LSP = defaultLSPConfig()
	return cfg
}

func mergeProviders(cfg *Config) {
	defaults := defaultProvidersMap()
	if cfg.Providers == nil {
		cfg.Providers = map[string]ProviderConfig{}
	}
	for name, pc := range defaults {
		if existing, ok := cfg.Providers[name]; ok {
			if existing.BaseURL == "" {
				existing.BaseURL = pc.BaseURL
			}
			cfg.Providers[name] = existing
		} else {
			cfg.Providers[name] = pc
		}
	}
}

func Load() (*Config, error) {
	if err := Ensure(); err != nil {
		return nil, fmt.Errorf("init opentmd dir: %w", err)
	}

	path, err := Path()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	cfg := DefaultConfig()
	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	mergeProviders(cfg)
	if cfg.Agent.MaxIterations <= 0 {
		cfg.Agent.MaxIterations = 10
	}
	cfg.LSP.Normalize()
	return cfg, nil
}

func Save(cfg *Config) error {
	dir, err := Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	path, err := Path()
	if err != nil {
		return err
	}

	data, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

func (c *Config) ActiveProvider() (string, ProviderConfig, error) {
	name := c.Default.Provider
	if name == "" {
		name = DefaultProvider
	}
	pc, ok := c.Providers[name]
	if !ok {
		return "", ProviderConfig{}, fmt.Errorf("provider %q not configured", name)
	}
	return name, pc, nil
}

func (c *Config) SetAPIKey(provider, key string) error {
	if _, ok := c.Providers[provider]; !ok {
		return fmt.Errorf("provider %q not found", provider)
	}
	pc := c.Providers[provider]
	pc.APIKey = key
	c.Providers[provider] = pc
	return Save(c)
}

func (c *Config) SetDefaultProvider(provider, model string) error {
	preset, ok := PresetByName(provider)
	if !ok {
		return fmt.Errorf("unknown provider %q", provider)
	}
	c.ApplyPreset(preset)
	if model != "" {
		c.Default.Model = model
	}
	return Save(c)
}

func (c *Config) MarkSetupCompleted() error {
	c.Setup.Completed = true
	return Save(c)
}

// NeedsWizard reports whether the interactive first-run wizard should run.
func (c *Config) NeedsWizard() bool {
	if c.Setup.Completed && c.HasConfiguredProvider() {
		return false
	}
	return !c.HasConfiguredProvider()
}

// HasConfiguredProvider checks if the active provider is ready to use.
func (c *Config) HasConfiguredProvider() bool {
	name, pc, err := c.ActiveProvider()
	if err != nil {
		return false
	}
	preset, ok := PresetByName(name)
	if ok && !preset.NeedAPIKey {
		return pc.BaseURL != ""
	}
	return pc.APIKey != ""
}
