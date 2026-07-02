package lsp

import (
	"time"

	"github.com/opentmd/opentmd-cli/internal/config"
)

type Config struct {
	Enabled         bool
	AutoDetect      bool
	SettleDelay     time.Duration
	WarmupFileLimit int
	Servers         map[string]config.LSPServerConfig
}

func ConfigFromApp(cfg *config.Config) Config {
	if cfg == nil {
		return DefaultConfig()
	}
	cfg.LSP.Normalize()
	return Config{
		Enabled:         cfg.LSP.Enabled,
		AutoDetect:      cfg.LSP.AutoDetect,
		SettleDelay:     cfg.LSP.SettleDelay(),
		WarmupFileLimit: cfg.LSP.WarmupFileLimit,
		Servers:         cfg.LSP.Servers,
	}
}

func DefaultConfig() Config {
	c := config.DefaultConfig()
	return ConfigFromApp(c)
}
