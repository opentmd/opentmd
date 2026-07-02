package daemon

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/opentmd/opentmd-cli/internal/config"
)

func TestHandleLSPStatus(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Providers["deepseek"] = config.ProviderConfig{
		BaseURL: "https://api.deepseek.com",
		APIKey:  "sk-test",
	}
	cfg.Setup.Completed = true

	s := New(cfg, 0)
	s.SetWorkDir(t.TempDir())

	req := httptest.NewRequest(http.MethodGet, "/lsp/status", nil)
	rec := httptest.NewRecorder()
	s.handleLSPStatus(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code: got %d body %s", rec.Code, rec.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["enabled"] != true {
		t.Fatalf("expected enabled=true, got %v", body["enabled"])
	}
	if _, ok := body["text"].(string); !ok {
		t.Fatalf("expected text field, got %v", body["text"])
	}
	if _, ok := body["servers"].([]any); !ok {
		t.Fatalf("expected servers array, got %T", body["servers"])
	}
}

func TestHandleLSPReload(t *testing.T) {
	s := New(config.DefaultConfig(), 0)

	req := httptest.NewRequest(http.MethodPost, "/lsp/reload", nil)
	rec := httptest.NewRecorder()
	s.handleLSPReload(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code: got %d", rec.Code)
	}
}

func TestHandleLSPStatusMethodNotAllowed(t *testing.T) {
	s := New(config.DefaultConfig(), 0)
	req := httptest.NewRequest(http.MethodPost, "/lsp/status", nil)
	rec := httptest.NewRecorder()
	s.handleLSPStatus(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}
