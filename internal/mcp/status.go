package mcp

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

type ServerStatusEntry struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	ToolCount int    `json:"tool_count,omitempty"`
	Error     string `json:"error,omitempty"`
	Command   string `json:"command,omitempty"`
}

type StatusReport struct {
	Servers []ServerStatusEntry `json:"servers"`
	Text    string            `json:"text"`
}

func ProbeStatus(ctx context.Context, workDir string) StatusReport {
	cfg, err := LoadConfig(workDir)
	if err != nil {
		return StatusReport{
			Text: fmt.Sprintf("MCP: failed to load config: %v", err),
		}
	}

	names := make([]string, 0, len(cfg.MCPServers))
	for name := range cfg.MCPServers {
		names = append(names, name)
	}
	sort.Strings(names)

	var servers []ServerStatusEntry
	for _, name := range names {
		sc := cfg.MCPServers[name]
		entry := ServerStatusEntry{Name: name, Command: sc.Command}
		switch {
		case sc.Disabled:
			entry.Status = "disabled"
		case sc.Command == "" && sc.URL == "":
			entry.Status = "skipped"
			entry.Error = "no command or url configured"
		case sc.Command == "":
			entry.Status = "skipped"
			entry.Error = "url transport not implemented"
		default:
			client, err := newStdioClient(sc.Command, sc.Args, sc.Env, workDir)
			if err != nil {
				entry.Status = "error"
				entry.Error = err.Error()
			} else {
				tools, err := client.listTools(ctx)
				_ = client.Close()
				if err != nil {
					entry.Status = "error"
					entry.Error = err.Error()
				} else {
					entry.Status = "connected"
					entry.ToolCount = len(tools)
				}
			}
		}
		servers = append(servers, entry)
	}

	report := StatusReport{Servers: servers, Text: FormatStatus(servers)}
	return report
}

func FormatStatus(servers []ServerStatusEntry) string {
	if len(servers) == 0 {
		return "MCP: no servers configured"
	}

	var connected, failed, other []string
	for _, s := range servers {
		label := s.Name
		switch s.Status {
		case "connected":
			connected = append(connected, fmt.Sprintf("%s (%d tools)", label, s.ToolCount))
		case "error":
			msg := s.Error
			if len(msg) > 80 {
				msg = msg[:80] + "…"
			}
			failed = append(failed, fmt.Sprintf("%s: %s", label, msg))
		default:
			other = append(other, fmt.Sprintf("%s [%s]", label, s.Status))
		}
	}
	sort.Strings(connected)
	sort.Strings(failed)
	sort.Strings(other)

	var lines []string
	lines = append(lines, fmt.Sprintf("MCP: %d server(s) configured", len(servers)))
	if len(connected) > 0 {
		lines = append(lines, "  connected: "+strings.Join(connected, ", "))
	}
	if len(failed) > 0 {
		lines = append(lines, "  failed: "+strings.Join(failed, ", "))
	}
	if len(other) > 0 {
		lines = append(lines, "  other: "+strings.Join(other, ", "))
	}
	return strings.Join(lines, "\n")
}

func (r *Registry) ToolCount() int {
	if r == nil {
		return 0
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.tools)
}

func (r *Registry) ServerNames() []string {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	seen := map[string]struct{}{}
	for _, t := range r.tools {
		seen[t.Server] = struct{}{}
	}
	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
