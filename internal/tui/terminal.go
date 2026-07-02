package tui

import (
	"os"
	"runtime"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
	"github.com/muesli/termenv"
)

// Capabilities describes what the active terminal can render reliably.
type Capabilities struct {
	Unicode bool
	Colors  bool
	Rounded bool
}

var termCaps Capabilities

// InitTerminal prepares stdout/stderr and selects glyph/color profiles.
func InitTerminal() Capabilities {
	cap := detectCapabilities()
	platformInit(&cap)
	refreshSymbols(cap)
	termCaps = cap

	switch {
	case !cap.Colors:
		lipgloss.SetColorProfile(termenv.Ascii)
	case !cap.Unicode:
		lipgloss.SetColorProfile(termenv.ANSI)
	}
	return cap
}

func detectCapabilities() Capabilities {
	unicode := !asciiMode()
	cap := Capabilities{
		Unicode: unicode,
		Colors:  isatty.IsTerminal(os.Stdout.Fd()) && os.Getenv("NO_COLOR") == "",
		Rounded: unicode && runtime.GOOS != "windows",
	}
	if runtime.GOOS == "windows" && (os.Getenv("WT_SESSION") != "" || os.Getenv("ConEmuANSI") == "ON") {
		cap.Rounded = true
	}
	return cap
}

func borderFor(cap Capabilities) lipgloss.Border {
	if cap.Rounded {
		return lipgloss.RoundedBorder()
	}
	return lipgloss.NormalBorder()
}
