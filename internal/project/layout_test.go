package project_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/opentmd/opentmd-cli/internal/project"
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

	if _, err := os.Stat(filepath.Join(dir, "OPENTMD.md")); err == nil {
		t.Fatal("OPENTMD.md should not be auto-created")
	}
}
