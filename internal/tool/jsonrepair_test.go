package tool

import (
	"encoding/json"
	"testing"
)

func TestRepairToolArgsTrailingComma(t *testing.T) {
	raw := `{"command":"ls",}`
	got := RepairToolArgs("bash", raw)
	if !json.Valid([]byte(got)) {
		t.Fatalf("expected valid json, got %q", got)
	}
}

func TestRepairToolArgsPassThroughValid(t *testing.T) {
	raw := `{"path":"src/main.go"}`
	got := RepairToolArgs("read_file", raw)
	if got != raw {
		t.Fatalf("expected unchanged, got %q", got)
	}
}
