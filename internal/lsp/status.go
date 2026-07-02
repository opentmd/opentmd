package lsp

import (
	"fmt"
	"sort"
	"strings"
)

type ServerEntry struct {
	Ext     string `json:"ext"`
	Command string `json:"command"`
	Active  bool   `json:"active"`
	InPath  bool   `json:"in_path"`
}

type Status struct {
	Enabled         bool          `json:"enabled"`
	AutoDetect      bool          `json:"auto_detect"`
	SettleDelayMs   int64         `json:"settle_delay_ms"`
	WarmupFileLimit int           `json:"warmup_file_limit"`
	Servers         []ServerEntry `json:"servers"`
	Text            string        `json:"text"`
}

func (m *Manager) Status() Status {
	st := Status{
		Enabled:         m.cfg.Enabled,
		AutoDetect:      m.cfg.AutoDetect,
		SettleDelayMs:   m.cfg.SettleDelay.Milliseconds(),
		WarmupFileLimit: m.cfg.WarmupFileLimit,
	}
	if !m.cfg.Enabled {
		st.Text = FormatStatus(st)
		return st
	}

	m.mu.Lock()
	active := make(map[string]struct{}, len(m.clients))
	for ext := range m.clients {
		active[ext] = struct{}{}
	}
	m.mu.Unlock()

	for _, ext := range m.reg.Extensions() {
		cfg, _ := m.reg.Get(ext)
		_, running := active[ext]
		st.Servers = append(st.Servers, ServerEntry{
			Ext:     ext,
			Command: cfg.Command,
			Active:  running,
			InPath:  commandAvailable(cfg.Command),
		})
	}
	st.Text = FormatStatus(st)
	return st
}

func FormatStatus(st Status) string {
	if !st.Enabled {
		return "LSP: disabled（在 config.toml [lsp] enabled = true 启用）"
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("LSP: enabled (auto_detect=%v, settle=%dms, warmup=%d)",
		st.AutoDetect, st.SettleDelayMs, st.WarmupFileLimit))

	if len(st.Servers) == 0 {
		lines = append(lines, "  无已配置语言服务器（启用 auto_detect 或配置 [lsp.servers.*]）")
		return strings.Join(lines, "\n")
	}

	var active, idle, missing []string
	for _, s := range st.Servers {
		label := fmt.Sprintf("%s (.%s)", s.Command, s.Ext)
		switch {
		case s.Active:
			active = append(active, label)
		case s.InPath:
			idle = append(idle, label)
		default:
			missing = append(missing, label)
		}
	}
	sort.Strings(active)
	sort.Strings(idle)
	sort.Strings(missing)

	if len(active) > 0 {
		lines = append(lines, "  已启动: "+strings.Join(active, ", "))
	}
	if len(idle) > 0 {
		lines = append(lines, "  未启动: "+strings.Join(idle, ", "))
	}
	if len(missing) > 0 {
		lines = append(lines, "  未安装: "+strings.Join(missing, ", "))
	}
	return strings.Join(lines, "\n")
}
