package mcp

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFormatStatusEmpty(t *testing.T) {
	if got := FormatStatus(nil); got != "MCP: no servers configured" {
		t.Fatalf("got %q", got)
	}
}

func TestProbeStatusDisabled(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	path := filepath.Join(dir, ".mcp.json")
	if err := os.WriteFile(path, []byte(`{
  "mcpServers": {
    "demo": {"command": "echo", "disabled": true}
  }
}`), 0o600); err != nil {
		t.Fatal(err)
	}

	report := ProbeStatus(context.Background(), dir)
	if len(report.Servers) != 1 {
		t.Fatalf("expected 1 server, got %d", len(report.Servers))
	}
	if report.Servers[0].Status != "disabled" {
		t.Fatalf("expected disabled, got %s", report.Servers[0].Status)
	}
	if !strings.Contains(report.Text, "other:") {
		t.Fatalf("expected other section, got %q", report.Text)
	}
}

func TestFormatStatusConnected(t *testing.T) {
	out := FormatStatus([]ServerStatusEntry{
		{Name: "fs", Status: "connected", ToolCount: 3},
		{Name: "git", Status: "error", Error: "spawn failed"},
	})
	if !strings.Contains(out, "connected: fs (3 tools)") {
		t.Fatalf("missing connected line: %q", out)
	}
	if !strings.Contains(out, "failed: git") {
		t.Fatalf("missing failed line: %q", out)
	}
}
