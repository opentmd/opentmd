package tui

import (
	"os"
	"runtime"
	"strings"
)

// sym holds terminal symbols selected at startup based on host capabilities.
var sym = defaultSymbols(true)

type termSymbols struct {
	Unicode  bool
	Spinner  []string
	Prompt   string
	Bullet   string
	Diamond  string
	Rule     string
	Ellipsis string
	Cursor   string
	Check    string
	Cross    string
	Warning  string
	Marker   string
	Arrow    string
	Up       string
	Down     string
}

func defaultSymbols(unicode bool) termSymbols {
	if unicode {
		return termSymbols{
			Unicode:  true,
			Spinner:  []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
			Prompt:   "❯ ",
			Bullet:   "∙ ",
			Diamond:  "◆ ",
			Rule:     "─",
			Ellipsis: "…",
			Cursor:   "▌",
			Check:    "✓",
			Cross:    "✗",
			Warning:  "⚠",
			Marker:   "●",
			Arrow:    "▸ ",
			Up:       "↑",
			Down:     "↓",
		}
	}
	return termSymbols{
		Unicode:  false,
		Spinner:  []string{"|", "/", "-", "\\"},
		Prompt:   "> ",
		Bullet:   "* ",
		Diamond:  "* ",
		Rule:     "-",
		Ellipsis: "...",
		Cursor:   "|",
		Check:    "+",
		Cross:    "x",
		Warning:  "!",
		Marker:   "*",
		Arrow:    "> ",
		Up:       "^",
		Down:     "v",
	}
}

func refreshSymbols(cap Capabilities) {
	sym = defaultSymbols(cap.Unicode)
}

func (s termSymbols) OK(msg string) string  { return s.Check + " " + msg }
func (s termSymbols) Err(msg string) string { return s.Cross + " " + msg }

func (s termSymbols) Normalize(text string) string {
	if s.Unicode {
		return text
	}
	repl := []struct{ from, to string }{
		{"❯", ">"},
		{"◆", "*"},
		{"∙", "*"},
		{"●", "*"},
		{"─", "-"},
		{"…", s.Ellipsis},
		{"✓", s.Check},
		{"✗", s.Cross},
		{"⚠", s.Warning},
		{"▶", ">"},
		{"▌", s.Cursor},
		{"▸", ">"},
		{"↑", s.Up},
		{"↓", s.Down},
		{"⠋", "|"},
		{"⠙", "/"},
		{"⠹", "-"},
		{"⠸", "\\"},
		{"⠼", "|"},
		{"⠴", "/"},
		{"⠦", "-"},
		{"⠧", "\\"},
		{"⠇", "|"},
		{"⠏", "/"},
	}
	for _, r := range repl {
		text = strings.ReplaceAll(text, r.from, r.to)
	}
	return text
}

// asciiMode reports whether the terminal cannot reliably render Unicode.
//
// Classic Windows CMD.exe lacks Braille spinner glyphs and several dingbats.
// Windows Terminal, ConEmu, VS Code, and Git Bash render Unicode correctly.
// Force ASCII mode on any platform with OPENTMD_ASCII=1.
func asciiMode() bool {
	if v := os.Getenv("OPENTMD_ASCII"); v == "1" || v == "true" || v == "yes" {
		return true
	}
	if os.Getenv("TERM") == "dumb" {
		return true
	}
	if runtime.GOOS != "windows" {
		return false
	}
	if os.Getenv("WT_SESSION") != "" {
		return false
	}
	if os.Getenv("ConEmuANSI") == "ON" {
		return false
	}
	if os.Getenv("TERM_PROGRAM") != "" {
		return false
	}
	if os.Getenv("MSYSTEM") != "" {
		return false
	}
	return true
}
