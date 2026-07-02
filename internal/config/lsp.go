package config

import "time"

type LSPServerConfig struct {
	Command     string   `toml:"command"`
	Args        []string `toml:"args"`
	RootMarkers []string `toml:"root_markers"`
}

type LSPConfig struct {
	Enabled                  bool                       `toml:"enabled"`
	AutoDetect               bool                       `toml:"auto_detect"`
	DiagnosticsSettleDelayMs int                        `toml:"diagnostics_settle_delay_ms"`
	WarmupFileLimit          int                        `toml:"warmup_file_limit"`
	Servers                  map[string]LSPServerConfig `toml:"servers"`
}

func defaultLSPConfig() LSPConfig {
	return LSPConfig{
		Enabled:                  true,
		AutoDetect:               true,
		DiagnosticsSettleDelayMs: 300,
		WarmupFileLimit:          40,
		Servers:                  map[string]LSPServerConfig{},
	}
}

func (c *LSPConfig) Normalize() {
	if c.DiagnosticsSettleDelayMs <= 0 {
		c.DiagnosticsSettleDelayMs = 300
	}
	if c.WarmupFileLimit <= 0 {
		c.WarmupFileLimit = 40
	}
	if c.Servers == nil {
		c.Servers = map[string]LSPServerConfig{}
	}
}

func (c *LSPConfig) SettleDelay() time.Duration {
	c.Normalize()
	return time.Duration(c.DiagnosticsSettleDelayMs) * time.Millisecond
}
