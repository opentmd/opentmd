package fileio

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const DefaultMaxBytes = 64 * 1024

func Read(path string, maxBytes int) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("path is empty")
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", err)
	}

	info, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("stat file: %w", err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("path is a directory")
	}
	if maxBytes <= 0 {
		maxBytes = DefaultMaxBytes
	}
	if info.Size() > int64(maxBytes) {
		return readTruncated(abs, maxBytes)
	}

	data, err := os.ReadFile(abs)
	if err != nil {
		return "", fmt.Errorf("read file: %w", err)
	}
	return string(data), nil
}

func readTruncated(path string, maxBytes int) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	buf := make([]byte, maxBytes)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("read file: %w", err)
	}
	content := string(buf[:n])
	return content + fmt.Sprintf("\n\n[truncated: showing first %d bytes]", maxBytes), nil
}

func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func IsReadable(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("path is a directory")
	}
	f, err := os.Open(abs)
	if err != nil {
		return err
	}
	return f.Close()
}
