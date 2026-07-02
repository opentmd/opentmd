package daemon

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/opentmd/opentmd-cli/internal/config"
)

func TestHandleMCPStatus(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	path := filepath.Join(dir, ".mcp.json")
	if err := os.WriteFile(path, []byte(`{
  "mcpServers": {
    "noop": {"command": "/nonexistent/mcp-server", "disabled": false}
  }
}`), 0o600); err != nil {
		t.Fatal(err)
	}

	s := New(config.DefaultConfig(), 0)
	s.SetWorkDir(dir)

	req := httptest.NewRequest(http.MethodGet, "/mcp/status", nil)
	rec := httptest.NewRecorder()
	s.handleMCPStatus(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code: got %d body %s", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	servers, ok := body["servers"].([]any)
	if !ok || len(servers) != 1 {
		t.Fatalf("expected servers array, got %v", body["servers"])
	}
	if _, ok := body["text"].(string); !ok {
		t.Fatalf("expected text field, got %v", body["text"])
	}
}

func TestHandleMCPStatusMethodNotAllowed(t *testing.T) {
	s := New(config.DefaultConfig(), 0)
	req := httptest.NewRequest(http.MethodPost, "/mcp/status", nil)
	rec := httptest.NewRecorder()
	s.handleMCPStatus(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}
