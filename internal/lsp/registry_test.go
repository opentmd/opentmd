package lsp

import (
	"testing"
)

func TestRegistryForFile(t *testing.T) {
	r := NewRegistry(true, nil)
	cfg, ok := r.ForFile("main.go")
	if !ok || cfg.Command != "gopls" {
		t.Fatalf("expected gopls, got %+v ok=%v", cfg, ok)
	}
	_, ok = r.ForFile("data.csv")
	if ok {
		t.Fatal("csv should have no server")
	}
}

func TestLanguageID(t *testing.T) {
	if languageID("go") != "go" {
		t.Fatal("go language id")
	}
	if languageID("tsx") != "typescriptreact" {
		t.Fatal("tsx language id")
	}
}
