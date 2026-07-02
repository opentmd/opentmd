package project

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/opentmd/opentmd-cli/internal/config"
)

const maxInstructionBytes = 1 << 20

var projectFiles = []string{
	".opentmd.md", "OPENTMD.md",
	".atomcode.md", "ATOMCODE.md",
	"CLAUDE.md", "claude.md",
}

func LoadContext(workDir string) string {
	var sections []string

	if cfgDir, err := config.Dir(); err == nil {
		if s := readFile(filepath.Join(cfgDir, "OPENTMD.md")); s != "" {
			sections = append(sections, "## Global Instructions\n"+s)
		}
	}

	for _, name := range projectFiles {
		if s := readFile(filepath.Join(workDir, name)); s != "" {
			sections = append(sections, "## Project Instructions ("+name+")\n"+s)
			break
		}
	}

	if s := readFile(filepath.Join(workDir, ".opentmd.user.md")); s != "" {
		sections = append(sections, "## User Instructions\n"+s)
	}

	if len(sections) == 0 {
		return ""
	}
	return strings.Join(sections, "\n\n")
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
