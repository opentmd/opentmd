package lsp

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestManagerCollectGoFile(t *testing.T) {
	if _, err := exec.LookPath("gopls"); err != nil {
		t.Skip("gopls not installed")
	}

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/test\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mainGo := filepath.Join(dir, "main.go")
	src := "package main\n\nfunc main() {\n\t_ = undefinedSymbol\n}\n"
	if err := os.WriteFile(mainGo, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	mgr := NewManager(dir, Config{Enabled: true, AutoDetect: true, SettleDelay: 800 * time.Millisecond, WarmupFileLimit: 5}, nil)
	defer mgr.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	out, err := mgr.Collect(ctx, "main.go", "error")
	if err != nil {
		t.Fatal(err)
	}
	if !containsAny(out, "undefined", "Undeclared", "diagnostic") {
		t.Fatalf("expected error diagnostic, got: %q", out)
	}
}

func containsAny(s string, parts ...string) bool {
	for _, p := range parts {
		if len(p) > 0 && stringContains(s, p) {
			return true
		}
	}
	return false
}

func stringContains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && indexSubstring(s, sub) >= 0)
}

func indexSubstring(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
