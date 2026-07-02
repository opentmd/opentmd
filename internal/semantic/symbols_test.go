package semantic

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListSymbolsGo(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.go")
	content := "package main\n\nfunc Hello() {}\n\ntype Widget struct {}\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	symbols, err := ListSymbols(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(symbols) < 2 {
		t.Fatalf("expected at least 2 symbols, got %d", len(symbols))
	}
}

func TestListSymbolsTreeSitterGo(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.go")
	content := "package main\n\nfunc Hello() {}\n\ntype Widget struct {}\n\nfunc (w Widget) Run() {}\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	symbols, err := ListSymbols(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(symbols) < 3 {
		t.Fatalf("expected at least 3 symbols, got %d: %+v", len(symbols), symbols)
	}
	names := map[string]bool{}
	for _, s := range symbols {
		names[s.Name] = true
	}
	for _, want := range []string{"Hello", "Widget", "Run"} {
		if !names[want] {
			t.Fatalf("missing symbol %q in %+v", want, symbols)
		}
	}
	for _, s := range symbols {
		if s.Name == "Hello" && s.Kind != "function_declaration" {
			t.Fatalf("expected tree-sitter kind function_declaration, got %q", s.Kind)
		}
	}
}
