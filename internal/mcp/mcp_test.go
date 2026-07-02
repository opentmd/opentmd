package mcp

import "testing"

func TestToolKey(t *testing.T) {
	key := ToolKey("filesystem", "read_file")
	if key != "mcp__filesystem__read_file" {
		t.Fatalf("got %q", key)
	}
}
