package project_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/opentmd/opentmd-cli/internal/project"
)

func TestLoadContextProjectOPENTMD(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	if err := os.WriteFile(filepath.Join(dir, "OPENTMD.md"), []byte("# Rules\n\nuse go test"), 0o644); err != nil {
		t.Fatal(err)
	}
	ctx := project.LoadContext(dir)
	if !strings.Contains(ctx, "Project Instructions (OPENTMD.md)") || !strings.Contains(ctx, "use go test") {
		t.Fatalf("unexpected context: %q", ctx)
	}
}

func TestLoadContextCaseInsensitive(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	if err := os.WriteFile(filepath.Join(dir, "opentmd.md"), []byte("lower case name"), 0o644); err != nil {
		t.Fatal(err)
	}
	ctx := project.LoadContext(dir)
	if !strings.Contains(ctx, "Project Instructions (opentmd.md)") || !strings.Contains(ctx, "lower case name") {
		t.Fatalf("unexpected context: %q", ctx)
	}
}

func TestLoadContextPrefersOPENTMDOverClaude(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("claude"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "OPENTMD.md"), []byte("opentmd"), 0o644); err != nil {
		t.Fatal(err)
	}
	ctx := project.LoadContext(dir)
	if !strings.Contains(ctx, "Project Instructions (OPENTMD.md)") || !strings.Contains(ctx, "opentmd") {
		t.Fatalf("expected OPENTMD.md to win: %q", ctx)
	}
	if strings.Contains(ctx, "claude") {
		t.Fatalf("should not load CLAUDE.md when OPENTMD.md exists: %q", ctx)
	}
}

func TestLoadContextClaudeFallback(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("from claude"), 0o644); err != nil {
		t.Fatal(err)
	}
	ctx := project.LoadContext(dir)
	if !strings.Contains(ctx, "Project Instructions (CLAUDE.md)") || !strings.Contains(ctx, "from claude") {
		t.Fatalf("unexpected context: %q", ctx)
	}
}

func TestLoadContextOptional(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	if ctx := project.LoadContext(dir); ctx != "" {
		t.Fatalf("expected empty context without instruction files, got %q", ctx)
	}
}

func TestLoadContextIgnoresLegacyDotOpentmd(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	if err := os.WriteFile(filepath.Join(dir, ".opentmd.md"), []byte("legacy"), 0o644); err != nil {
		t.Fatal(err)
	}
	if ctx := project.LoadContext(dir); ctx != "" {
		t.Fatalf(".opentmd.md should be ignored, got %q", ctx)
	}
}