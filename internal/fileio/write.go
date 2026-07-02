package fileio

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func Write(path string, content string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}

func Edit(path, oldStr, newStr string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read file: %w", err)
	}
	content := string(data)
	if !strings.Contains(content, oldStr) {
		return "", fmt.Errorf("old_string not found in %s", path)
	}
	updated := strings.Replace(content, oldStr, newStr, 1)
	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}
	return diffPreview(oldStr, newStr), nil
}

func diffPreview(oldStr, newStr string) string {
	return fmt.Sprintf("applied edit:\n---\n-%s\n+%s", truncateLines(oldStr, 8), truncateLines(newStr, 8))
}

func truncateLines(s string, max int) string {
	lines := strings.Split(s, "\n")
	if len(lines) <= max {
		return s
	}
	return strings.Join(lines[:max], "\n") + "\n..."
}
