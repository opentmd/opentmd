package tui

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/opentmd/opentmd-cli/internal/config"
)

const inputHistoryMax = 1000

type inputHistory struct {
	path    string
	entries []string
}

func loadInputHistory() (*inputHistory, error) {
	dir, err := config.Dir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(dir, "input-history.jsonl")
	h := &inputHistory{path: path}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return h, nil
		}
		return nil, err
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var entry struct {
			Text string `json:"text"`
		}
		if json.Unmarshal([]byte(line), &entry) == nil && entry.Text != "" {
			h.entries = append(h.entries, entry.Text)
			continue
		}
		var plain string
		if json.Unmarshal([]byte(line), &plain) == nil && plain != "" {
			h.entries = append(h.entries, plain)
			continue
		}
		h.entries = append(h.entries, line)
	}
	if len(h.entries) > inputHistoryMax {
		h.entries = h.entries[len(h.entries)-inputHistoryMax:]
	}
	return h, nil
}

func (h *inputHistory) add(text string) error {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	if len(h.entries) > 0 && h.entries[len(h.entries)-1] == text {
		return nil
	}
	h.entries = append(h.entries, text)
	if len(h.entries) > inputHistoryMax {
		h.entries = h.entries[len(h.entries)-inputHistoryMax:]
	}
	return h.persist()
}

func (h *inputHistory) persist() error {
	if err := os.MkdirAll(filepath.Dir(h.path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(h.path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	for _, e := range h.entries {
		b, err := json.Marshal(struct{ Text string }{Text: e})
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, string(b)); err != nil {
			return err
		}
	}
	return w.Flush()
}

func (h *inputHistory) prev(stored string, index int) (string, int) {
	if len(h.entries) == 0 {
		return stored, -1
	}
	if index < 0 {
		index = len(h.entries)
	}
	index--
	if index < 0 {
		index = 0
	}
	return h.entries[index], index
}

func (h *inputHistory) next(index int) (string, int) {
	if index < 0 || index >= len(h.entries)-1 {
		return "", -1
	}
	index++
	return h.entries[index], index
}
