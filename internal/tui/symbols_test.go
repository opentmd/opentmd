package tui

import "testing"

func TestAsciiModeEnv(t *testing.T) {
	t.Setenv("OPENTMD_ASCII", "1")
	if !asciiMode() {
		t.Fatal("OPENTMD_ASCII=1 should enable ascii mode")
	}
}

func TestSymbolsNormalizeASCII(t *testing.T) {
	s := defaultSymbols(false)
	out := s.Normalize("✓ ok ❯ prompt ● bullet ─ rule …")
	if out != "+ ok > prompt * bullet - rule ..." {
		t.Fatalf("unexpected normalize: %q", out)
	}
}

func TestRefreshSymbols(t *testing.T) {
	refreshSymbols(Capabilities{Unicode: false})
	if sym.Prompt != "> " {
		t.Fatalf("expected ascii prompt, got %q", sym.Prompt)
	}
	refreshSymbols(Capabilities{Unicode: true})
	if sym.Prompt != "❯ " {
		t.Fatalf("expected unicode prompt, got %q", sym.Prompt)
	}
}
