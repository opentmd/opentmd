package grep

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/opentmd/opentmd/internal/pathutil"
)

type Options struct {
	Pattern    string
	Root       string
	MaxResults int
	Context    int
}

type Match struct {
	File    string
	Line    int
	Content string
}

func Search(opt Options) (string, error) {
	if opt.MaxResults <= 0 {
		opt.MaxResults = 50
	}
	if opt.Context <= 0 {
		opt.Context = 3
	}
	if opt.Context > 10 {
		opt.Context = 10
	}

	re, literal, err := compilePattern(opt.Pattern)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(opt.Root)
	if err != nil {
		return "", fmt.Errorf("path not found: %w", err)
	}

	var matches []Match
	if !info.IsDir() {
		m, err := searchFile(opt.Root, opt.Root, re, literal, opt.Context, opt.MaxResults)
		if err != nil {
			return "", err
		}
		matches = m
	} else {
		_ = filepath.WalkDir(opt.Root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				if path != opt.Root && pathutil.ShouldSkipDir(d.Name()) {
					return filepath.SkipDir
				}
				return nil
			}
			if len(matches) >= opt.MaxResults {
				return fs.SkipAll
			}
			fileMatches, err := searchFile(path, opt.Root, re, literal, opt.Context, opt.MaxResults-len(matches))
			if err != nil {
				return nil
			}
			matches = append(matches, fileMatches...)
			return nil
		})
	}

	if len(matches) == 0 {
		return "no matches found", nil
	}

	var sb strings.Builder
	for _, m := range matches {
		sb.WriteString(fmt.Sprintf("%s:%d: %s\n", m.File, m.Line, m.Content))
	}
	if len(matches) >= opt.MaxResults {
		sb.WriteString(fmt.Sprintf("\n(truncated at %d results)", opt.MaxResults))
	}
	return sb.String(), nil
}

func compilePattern(pattern string) (*regexp.Regexp, string, error) {
	hasUpper := false
	for _, r := range pattern {
		if r >= 'A' && r <= 'Z' {
			hasUpper = true
			break
		}
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, pattern, nil
	}
	if !hasUpper {
		re, err = regexp.Compile("(?i:" + pattern + ")")
		if err != nil {
			re, _ = regexp.Compile(pattern)
		}
	}
	return re, "", nil
}

func searchFile(path, root string, re *regexp.Regexp, literal string, ctx, limit int) ([]Match, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	rel, _ := filepath.Rel(root, path)
	var out []Match
	for i, line := range lines {
		matched := false
		if re != nil {
			matched = re.MatchString(line)
		} else if literal != "" {
			matched = strings.Contains(line, literal)
		}
		if !matched {
			continue
		}
		start := i - ctx
		if start < 0 {
			start = 0
		}
		end := i + ctx + 1
		if end > len(lines) {
			end = len(lines)
		}
		block := strings.Join(lines[start:end], "\n")
		out = append(out, Match{File: rel, Line: i + 1, Content: block})
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}
