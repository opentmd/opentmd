package project

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/opentmd/opentmd-cli/internal/config"
)

const maxInstructionBytes = 1 << 20

// Instruction file names in priority order (Claude-compatible fallbacks last).
var instructionNames = []string{
	"OPENTMD.md",
	"CLAUDE.md",
	".atomcode.md",
	"ATOMCODE.md",
}

var instructionNameRank = map[string]int{
	"opentmd.md":   0,
	"claude.md":    1,
	".atomcode.md": 2,
	"atomcode.md":  3,
}

func LoadContext(workDir string) string {
	var sections []string

	if cfgDir, err := config.Dir(); err == nil {
		if name, content := findInstructions(cfgDir); content != "" {
			sections = append(sections, "## Global Instructions ("+name+")\n"+content)
		}
	}

	if name, content := findInstructions(workDir); content != "" {
		sections = append(sections, "## Project Instructions ("+name+")\n"+content)
	}

	if s := readFile(filepath.Join(workDir, ".opentmd.user.md")); s != "" {
		sections = append(sections, "## User Instructions\n"+s)
	}

	if len(sections) == 0 {
		return ""
	}
	return strings.Join(sections, "\n\n")
}

// findInstructions locates an instruction markdown file in dir.
// OPENTMD.md is preferred; names are matched case-insensitively on disk.
// Returns the file name (not full path) and trimmed content.
func findInstructions(dir string) (string, string) {
	for _, name := range instructionNames {
		path := filepath.Join(dir, name)
		if content := readFile(path); content != "" {
			return name, content
		}
	}

	path, name := findInstructionsCaseInsensitive(dir)
	if path == "" {
		return "", ""
	}
	return name, readFile(path)
}

func findInstructionsCaseInsensitive(dir string) (path, name string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", ""
	}

	bestRank := -1
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		base := entry.Name()
		rank, ok := instructionNameRank[strings.ToLower(base)]
		if !ok {
			continue
		}
		if bestRank == -1 || rank < bestRank {
			bestRank = rank
			name = base
			path = filepath.Join(dir, base)
		}
	}
	return path, name
}

func DetectStack(workDir string) string {
	checks := []struct {
		file  string
		label string
	}{
		{"go.mod", "Go"},
		{"package.json", "Node.js"},
		{"Cargo.toml", "Rust"},
		{"pyproject.toml", "Python"},
		{"requirements.txt", "Python"},
		{"pom.xml", "Java"},
	}
	var found []string
	for _, c := range checks {
		if _, err := os.Stat(filepath.Join(workDir, c.file)); err == nil {
			found = append(found, c.label)
		}
	}
	if len(found) == 0 {
		return ""
	}
	return "Detected stack: " + strings.Join(found, ", ")
}

func readFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	if len(data) > maxInstructionBytes {
		data = data[:maxInstructionBytes]
	}
	s := strings.TrimSpace(string(data))
	return s
}
