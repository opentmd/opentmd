package tui

import (
	"os"
	"runtime"
)

// useAltScreen decides whether bubbletea uses the alternate screen buffer.
// SSH sessions without a display cannot use X11/wl-clipboard and often block
// OSC52; disabling alt screen keeps terminal scrollback so users can select
// output with the mouse. Legacy Windows CMD also keeps scrollback without alt screen.
func useAltScreen() bool {
	switch os.Getenv("OPENTMD_NO_ALT_SCREEN") {
	case "1", "true", "yes":
		return false
	case "0", "false", "no":
		return true
	}
	if os.Getenv("OPENTMD_ALT_SCREEN") == "1" {
		return true
	}
	if os.Getenv("SSH_CONNECTION") != "" &&
		os.Getenv("DISPLAY") == "" &&
		os.Getenv("WAYLAND_DISPLAY") == "" {
		return false
	}
	if runtime.GOOS == "windows" &&
		os.Getenv("WT_SESSION") == "" &&
		os.Getenv("ConEmuANSI") != "ON" {
		return false
	}
	return true
}
