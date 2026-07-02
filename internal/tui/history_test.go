package tui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInputHistoryAddAndBrowse(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	h, err := loadInputHistory()
	if err != nil {
		t.Fatal(err)
	}
	if err := h.add("hello"); err != nil {
		t.Fatal(err)
	}
	if err := h.add("world"); err != nil {
		t.Fatal(err)
	}
	entry, idx := h.prev("", -1)
	if entry != "world" || idx != 1 {
		t.Fatalf("prev: got %q %d", entry, idx)
	}
	entry, idx = h.prev("", idx)
	if entry != "hello" || idx != 0 {
		t.Fatalf("prev2: got %q %d", entry, idx)
	}
	entry, idx = h.next(0)
	if entry != "world" || idx != 1 {
		t.Fatalf("next: got %q %d", entry, idx)
	}

	path := filepath.Join(dir, ".opentmd", "input-history.jsonl")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected persisted history: %v", err)
	}
}

func TestCycleSlashSelectionWrap(t *testing.T) {
	if got := cycleSlashSelection(0, 3, true); got != 2 {
		t.Fatalf("expected wrap to 2, got %d", got)
	}
}
