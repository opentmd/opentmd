package project_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/opentmd/opentmd-cli/internal/config"
	"github.com/opentmd/opentmd-cli/internal/project"
)

func TestInitWorkspacePersistsDefaultWorkdir(t *testing.T) {
	home := t.TempDir()
	projectDir := filepath.Join(home, "my-project")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)

	cfg := config.DefaultConfig()
	if err := project.InitWorkspace(projectDir, cfg); err != nil {
		t.Fatalf("init workspace: %v", err)
	}
	if cfg.DefaultWorkdir != projectDir {
		t.Fatalf("expected default_workdir %q, got %q", projectDir, cfg.DefaultWorkdir)
	}

	loaded, err := config.Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if loaded.DefaultWorkdir != projectDir {
		t.Fatalf("expected persisted default_workdir %q, got %q", projectDir, loaded.DefaultWorkdir)
	}
}

func TestResolveWorkDirUsesCWD(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	got, err := project.ResolveWorkDir("")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	want, _ := filepath.Abs(dir)
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestResolveWorkDirOverride(t *testing.T) {
	base := t.TempDir()
	sub := filepath.Join(base, "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(base)

	got, err := project.ResolveWorkDir(sub)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	want, _ := filepath.Abs(sub)
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
