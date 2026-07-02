package tui

import "testing"

func TestUseAltScreenExplicitOverride(t *testing.T) {
	t.Setenv("SSH_CONNECTION", "1.2.3.4 12345 5.6.7.8 22")
	t.Setenv("DISPLAY", "")
	t.Setenv("WAYLAND_DISPLAY", "")

	t.Setenv("OPENTMD_ALT_SCREEN", "1")
	if !useAltScreen() {
		t.Fatal("OPENTMD_ALT_SCREEN=1 should force alt screen on")
	}

	t.Setenv("OPENTMD_ALT_SCREEN", "")
	t.Setenv("OPENTMD_NO_ALT_SCREEN", "1")
	if useAltScreen() {
		t.Fatal("OPENTMD_NO_ALT_SCREEN=1 should force alt screen off")
	}
}
