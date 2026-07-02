package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCopyViaOSC52GeneratesSequence(t *testing.T) {
	t.Setenv("TMUX", "")
	t.Setenv("STY", "")

	if os.Getenv("CI") != "" && clipboardUnsupported() {
		t.Skip("no clipboard backend in CI")
	}
	if err := copyViaOSC52("hello opentmd"); err != nil {
		t.Fatalf("copyViaOSC52: %v", err)
	}
}

func TestLastNonSystemText(t *testing.T) {
	m := Model{messages: []displayMsg{
		{kind: "user", text: "hi"},
		{kind: "tool", text: "Reading read_file(...)"},
		{kind: "system", text: "✓ done"},
	}}
	if got := m.lastNonSystemText(); got != "Reading read_file(...)" {
		t.Fatalf("expected tool text, got %q", got)
	}
}

func TestHandleCopyCommandFallsBackToToolOutput(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	m := Model{messages: []displayMsg{
		{kind: "user", text: "list files"},
		{kind: "assistant", text: ""},
		{kind: "tool", text: "Reading list_directory({})"},
	}}
	m.handleCopyCommand(nil)
	last := m.messages[len(m.messages)-1].text
	if !strings.Contains(last, "已复制") && !strings.Contains(last, "已写入文件") {
		t.Fatalf("expected copy result system message, got %q", last)
	}
	fallback := filepath.Join(dir, ".opentmd", "last-copy.txt")
	if _, err := os.Stat(fallback); err != nil {
		t.Fatalf("expected fallback file: %v", err)
	}
}

func TestUseAltScreenDisabledOnSSHWithoutDisplay(t *testing.T) {
	t.Setenv("OPENTMD_NO_ALT_SCREEN", "")
	t.Setenv("OPENTMD_ALT_SCREEN", "")
	t.Setenv("SSH_CONNECTION", "1.2.3.4 12345 5.6.7.8 22")
	t.Setenv("DISPLAY", "")
	t.Setenv("WAYLAND_DISPLAY", "")
	if useAltScreen() {
		t.Fatal("expected alt screen disabled on SSH without display")
	}
}

func TestUseAltScreenEnabledByDefault(t *testing.T) {
	t.Setenv("OPENTMD_NO_ALT_SCREEN", "")
	t.Setenv("OPENTMD_ALT_SCREEN", "")
	t.Setenv("SSH_CONNECTION", "")
	t.Setenv("DISPLAY", ":0")
	if !useAltScreen() {
		t.Fatal("expected alt screen enabled with DISPLAY")
	}
}

func clipboardUnsupported() bool {
	return os.Getenv("DISPLAY") == "" && os.Getenv("WAYLAND_DISPLAY") == ""
}
