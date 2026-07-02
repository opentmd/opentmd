package pathutil

import (
	"fmt"
	"path/filepath"
	"strings"
)

var SkipDirs = map[string]bool{
	".git": true, "node_modules": true, "target": true, "vendor": true,
	"__pycache__": true, ".opentmd": true, ".atomcode": true,
	"dist": true, "build": true, ".next": true,
}

func Resolve(workDir, p string) (string, error) {
	if workDir == "" {
		var err error
		workDir, err = filepath.Abs(".")
		if err != nil {
			return "", err
		}
	}
	workDir, err := filepath.Abs(workDir)
	if err != nil {
		return "", err
	}

	if p == "" || p == "." {
		return workDir, nil
	}

	var abs string
	if filepath.IsAbs(p) {
		abs = filepath.Clean(p)
	} else {
		abs = filepath.Clean(filepath.Join(workDir, p))
	}

	rel, err := filepath.Rel(workDir, abs)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("path %q is outside workspace %q", p, workDir)
	}
	return abs, nil
}

func ShouldSkipDir(name string) bool {
	return SkipDirs[name]
}
