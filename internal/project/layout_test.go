package project_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/opentmd/opentmd/internal/project"
)

func TestEnsureCreatesProjectLayout(t *testing.T) {
	dir := t.TempDir()

	if err := project.Ensure(dir); err != nil {
		t.Fatalf("ensure: %v", err)
	}

	root := project.Dir(dir)
	for _, name := range []string{"skills", "local", "skills/example/SKILL.md", "mcp.json", "hooks.json", "local/.gitkeep"} {
		path := filepath.Join(root, name)
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected %s to exist: %v", path, err)
		}
	}

	if _, err := os.Stat(filepath.Join(dir, ".opentmd.md")); err != nil {
		t.Fatalf("expected .opentmd.md starter: %v", err)
	}
}

func TestEnsureDoesNotOverwriteExistingInstructions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".opentmd.md")
	const content = "# Custom\n\nkeep me"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := project.Ensure(dir); err != nil {
		t.Fatalf("ensure: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != content {
		t.Fatalf("instructions overwritten: %q", data)
	}
}
