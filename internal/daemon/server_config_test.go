package daemon

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/opentmd/opentmd/internal/config"
)

func TestHandleConfigReload(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	opentmdDir := filepath.Join(dir, ".opentmd")
	if err := os.MkdirAll(opentmdDir, 0o755); err != nil {
		t.Fatal(err)
	}
	cfgPath := filepath.Join(opentmdDir, "config.toml")
	if err := os.WriteFile(cfgPath, []byte(`
[default]
provider = "deepseek"
model = "deepseek-chat"

[providers.deepseek]
base_url = "https://api.deepseek.com"
api_key = "sk-test"

[setup]
completed = true
`), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	s := New(cfg, 0)

	req := httptest.NewRequest(http.MethodPost, "/config/reload", nil)
	rec := httptest.NewRecorder()
	s.handleConfigReload(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code: got %d body %s", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["status"] != "ok" {
		t.Fatalf("expected ok, got %v", body["status"])
	}
}

func TestHandleConfigMethodNotAllowed(t *testing.T) {
	s := New(config.DefaultConfig(), 0)
	req := httptest.NewRequest(http.MethodPost, "/config", nil)
	rec := httptest.NewRecorder()
	s.handleConfig(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}
