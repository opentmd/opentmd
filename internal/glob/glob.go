package glob

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/opentmd/opentmd-cli/internal/pathutil"
)

type Options struct {
	Pattern string
	Root    string
	Limit   int
}

func Find(opt Options) (string, error) {
	if opt.Limit <= 0 {
		opt.Limit = 200
	}
	pattern := filepath.ToSlash(opt.Pattern)
	searchRoot, namePattern := splitPattern(opt.Root, pattern)

	if _, err := os.Stat(searchRoot); err != nil {
		return "", fmt.Errorf("directory not found: %s", searchRoot)
	}

	var matches []string
	_ = filepath.WalkDir(searchRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if path != searchRoot && pathutil.ShouldSkipDir(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if len(matches) >= opt.Limit {
			return fs.SkipAll
		}
		rel, err := filepath.Rel(opt.Root, path)
		if err != nil {
			rel = path
		}
		rel = filepath.ToSlash(rel)
		if matchName(namePattern, filepath.Base(path)) || matchPath(pattern, rel) {
			matches = append(matches, rel)
		}
		return nil
	})

	if len(matches) == 0 {
		return "no files matched", nil
	}
	result := strings.Join(matches, "\n")
	if len(matches) >= opt.Limit {
		result += fmt.Sprintf("\n(truncated at %d files)", opt.Limit)
	}
	return result, nil
}

func splitPattern(root, pattern string) (searchRoot, namePattern string) {
	searchRoot = root
	namePattern = pattern
	pattern = strings.TrimPrefix(pattern, "./")
	if strings.Contains(pattern, "/") {
		parts := strings.Split(pattern, "/")
		var dirParts []string
		for i, p := range parts {
			if p == "**" {
				namePattern = strings.Join(parts[i+1:], "/")
				break
			}
			if strings.Contains(p, "*") || strings.Contains(p, "?") {
				namePattern = strings.Join(parts[i:], "/")
				break
			}
			dirParts = append(dirParts, p)
		}
		if len(dirParts) > 0 {
			searchRoot = filepath.Join(root, filepath.Join(dirParts...))
		}
	}
	if namePattern == "" {
		namePattern = "*"
	}
	return searchRoot, filepath.Base(namePattern)
}

func matchName(pattern, name string) bool {
	ok, _ := filepath.Match(pattern, name)
	return ok
}

func matchPath(fullPattern, rel string) bool {
	if strings.Contains(fullPattern, "**") {
		suffix := strings.TrimPrefix(fullPattern, "**/")
		if suffix != fullPattern {
			ok, _ := filepath.Match(suffix, filepath.Base(rel))
			if ok {
				return true
			}
			ok, _ = filepath.Match(suffix, rel)
			return ok
		}
	}
	ok, _ := filepath.Match(fullPattern, rel)
	return ok
}
