// Package memory provides persistent memory.md storage.
//
// Two stores:
//   - global: ~/.opentmd/memory.md  (user preferences across projects)
//   - project: <root>/.opentmd/memory.md  (project-specific facts)
//
// Entries are plain "- " bullet lines — human-editable, git-diffable.
// Merged output is injected into the system prompt at session start.
package memory

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/opentmd/opentmd-cli/internal/config"
)

const (
	maxMemoryFileSize = 64 * 1024 // 64 KB tail-read cap
	defaultCharLimit  = 4000      // max chars in merged prompt output
	memoryHeader      = "=== MEMORY ==="
)

// Store manages a single memory.md file.
type Store struct {
	path string
}

// New opens a memory store at the given path.
func New(path string) *Store {
	return &Store{path: path}
}

// Global returns the global memory store (~/.opentmd/memory.md).
func Global() *Store {
	dir, err := config.Dir()
	if err != nil {
		return New("")
	}
	return New(filepath.Join(dir, "memory.md"))
}

// Project returns the project-level memory store (<root>/.opentmd/memory.md).
func Project(workDir string) *Store {
	if workDir == "" {
		return New("")
	}
	return New(filepath.Join(workDir, ".opentmd", "memory.md"))
}

// Path returns the underlying file path.
func (s *Store) Path() string {
	return s.path
}

// Valid returns true when the store has a usable path.
func (s *Store) Valid() bool {
	return s.path != ""
}

// Load reads all memory entries from the file.
// Returns nil (not an error) when the file doesn't exist.
func (s *Store) Load() []string {
	if !s.Valid() {
		return nil
	}
	content := s.readFile()
	if content == "" {
		return nil
	}
	var entries []string
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if rest, ok := strings.CutPrefix(trimmed, "- "); ok {
			entries = append(entries, rest)
		}
	}
	return entries
}

// Append adds a new "- <content>" line to the memory file.
// Creates parent directories if needed.
func (s *Store) Append(content string) error {
	if !s.Valid() {
		return fmt.Errorf("memory store: no path set")
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("create memory dir: %w", err)
	}

	content = strings.TrimSpace(content)
	if content == "" {
		return nil
	}

	// Check if we need a leading newline
	needsNewline := false
	if existing, err := os.ReadFile(s.path); err == nil && len(existing) > 0 && existing[len(existing)-1] != '\n' {
		needsNewline = true
	}

	f, err := os.OpenFile(s.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open memory file: %w", err)
	}
	defer f.Close()

	if needsNewline {
		if _, err := fmt.Fprintln(f); err != nil {
			return err
		}
	}
	_, err = fmt.Fprintf(f, "- %s\n", content)
	return err
}

// RemoveMatching removes all entries whose text contains the keyword (case-insensitive).
// Returns the removed entries.
func (s *Store) RemoveMatching(keyword string) ([]string, error) {
	if !s.Valid() {
		return nil, fmt.Errorf("memory store: no path set")
	}
	content, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	keyword = strings.ToLower(keyword)
	var kept []string
	var removed []string

	for _, line := range strings.Split(string(content), "\n") {
		trimmed := strings.TrimSpace(line)
		if rest, ok := strings.CutPrefix(trimmed, "- "); ok && strings.Contains(strings.ToLower(rest), keyword) {
			removed = append(removed, rest)
		} else {
			kept = append(kept, line)
		}
	}

	if len(removed) == 0 {
		return nil, nil
	}

	out := strings.Join(kept, "\n")
	if !strings.HasSuffix(out, "\n") {
		out += "\n"
	}
	if err := os.WriteFile(s.path, []byte(out), 0o644); err != nil {
		return nil, err
	}
	return removed, nil
}

// FindMatching returns entries containing the keyword (case-insensitive).
func (s *Store) FindMatching(keyword string) []string {
	entries := s.Load()
	if len(entries) == 0 {
		return nil
	}
	keyword = strings.ToLower(keyword)
	var out []string
	for _, e := range entries {
		if strings.Contains(strings.ToLower(e), keyword) {
			out = append(out, e)
		}
	}
	return out
}

// MergedForPrompt combines global and project memory into a system-prompt block.
// Returns empty string when both stores are empty.
func MergedForPrompt(global, project *Store, projectName string) string {
	globalEntries := global.Load()
	projectEntries := project.Load()

	if len(globalEntries) == 0 && len(projectEntries) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(memoryHeader)
	sb.WriteString("\nThe user has asked you to remember these facts and preferences:\n")

	if len(globalEntries) > 0 {
		sb.WriteString("\n[Global]\n")
		for _, e := range globalEntries {
			fmt.Fprintf(&sb, "- %s\n", e)
		}
	}

	if len(projectEntries) > 0 {
		fmt.Fprintf(&sb, "\n[Project: %s]\n", projectName)
		for _, e := range projectEntries {
			fmt.Fprintf(&sb, "- %s\n", e)
		}
	}

	result := sb.String()
	if len([]rune(result)) > defaultCharLimit {
		runes := []rune(result)
		result = string(runes[:defaultCharLimit]) + "\n[...truncated, run /memory to review]"
	}
	return result
}

func (s *Store) readFile() string {
	if !s.Valid() {
		return ""
	}
	info, err := os.Stat(s.path)
	if err != nil {
		return ""
	}
	if info.Size() > maxMemoryFileSize {
		data, err := os.ReadFile(s.path)
		if err != nil {
			return ""
		}
		start := len(data) - int(maxMemoryFileSize)
		// Scan forward to next newline to avoid splitting UTF-8 chars
		if start < 0 {
			start = 0
		}
		if idx := indexOf(data[start:], '\n'); idx >= 0 && start+idx+1 < len(data) {
			start = start + idx + 1
		}
		return string(data[start:])
	}
	data, err := os.ReadFile(s.path)
	if err != nil {
		return ""
	}
	return string(data)
}

func indexOf(data []byte, b byte) int {
	for i, v := range data {
		if v == b {
			return i
		}
	}
	return -1
}
