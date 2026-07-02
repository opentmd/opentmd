package project

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/opentmd/opentmd-cli/internal/config"
)

// ResolveWorkDir returns the absolute workspace path. When override is non-empty,
// the process chdirs there first (same as opentmd -C).
func ResolveWorkDir(override string) (string, error) {
	if override != "" {
		if err := os.Chdir(override); err != nil {
			return "", fmt.Errorf("chdir: %w", err)
		}
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getwd: %w", err)
	}
	abs, err := filepath.Abs(wd)
	if err != nil {
		return "", fmt.Errorf("abs path: %w", err)
	}
	return filepath.Clean(abs), nil
}

// InitWorkspace creates the project .opentmd layout and persists workDir as
// default_workdir in the global config.
func InitWorkspace(workDir string, cfg *config.Config) error {
	if err := Ensure(workDir); err != nil {
		return err
	}
	if cfg == nil {
		return nil
	}
	if cfg.DefaultWorkdir == workDir {
		return nil
	}
	cfg.DefaultWorkdir = workDir
	return config.Save(cfg)
}
