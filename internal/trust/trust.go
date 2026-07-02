package trust

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/opentmd/opentmd/internal/config"
)

var ErrDeclined = errors.New("workspace trust declined")

type storeFile struct {
	Workspaces []string `json:"workspaces"`
}

type Options struct {
	Skip      bool // OPENTMD_SKIP_TRUST=1
	AutoTrust bool // --trust
}

var fileMu sync.Mutex

func Normalize(workDir string) (string, error) {
	if workDir == "" {
		return "", fmt.Errorf("empty workspace path")
	}
	abs, err := filepath.Abs(workDir)
	if err != nil {
		return "", fmt.Errorf("abs path: %w", err)
	}
	return filepath.Clean(abs), nil
}

func storePath() (string, error) {
	dir, err := config.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "trusted-workspaces.json"), nil
}

func loadStore() (*storeFile, error) {
	path, err := storePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &storeFile{}, nil
		}
		return nil, fmt.Errorf("read trust store: %w", err)
	}
	var sf storeFile
	if err := json.Unmarshal(data, &sf); err != nil {
		return nil, fmt.Errorf("parse trust store: %w", err)
	}
	return &sf, nil
}

func saveStore(sf *storeFile) error {
	path, err := storePath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(sf, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal trust store: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write trust store: %w", err)
	}
	return nil
}

func IsTrusted(workDir string) (bool, error) {
	norm, err := Normalize(workDir)
	if err != nil {
		return false, err
	}
	sf, err := loadStore()
	if err != nil {
		return false, err
	}
	for _, p := range sf.Workspaces {
		if p == norm {
			return true, nil
		}
	}
	return false, nil
}

func Trust(workDir string) error {
	norm, err := Normalize(workDir)
	if err != nil {
		return err
	}
	fileMu.Lock()
	defer fileMu.Unlock()

	sf, err := loadStore()
	if err != nil {
		return err
	}
	for _, p := range sf.Workspaces {
		if p == norm {
			return nil
		}
	}
	sf.Workspaces = append(sf.Workspaces, norm)
	return saveStore(sf)
}

func Untrust(workDir string) error {
	norm, err := Normalize(workDir)
	if err != nil {
		return err
	}
	fileMu.Lock()
	defer fileMu.Unlock()

	sf, err := loadStore()
	if err != nil {
		return err
	}
	var kept []string
	found := false
	for _, p := range sf.Workspaces {
		if p == norm {
			found = true
			continue
		}
		kept = append(kept, p)
	}
	if !found {
		return nil
	}
	sf.Workspaces = kept
	return saveStore(sf)
}

// Resolve returns whether the workspace is trusted and whether a prompt is needed.
func Resolve(workDir string, opts Options) (trusted bool, needPrompt bool, err error) {
	if opts.Skip {
		return true, false, nil
	}
	if opts.AutoTrust {
		if err := Trust(workDir); err != nil {
			return false, false, err
		}
		return true, false, nil
	}
	ok, err := IsTrusted(workDir)
	if err != nil {
		return false, false, err
	}
	if ok {
		return true, false, nil
	}
	return false, true, nil
}
