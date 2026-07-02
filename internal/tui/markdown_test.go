package tui

import (
	"strings"
	"testing"
)

func TestRenderMarkdownLiteCodeBlock(t *testing.T) {
	s := loadTheme("dark").styles()
	in := "hello\n```go\nfmt.Println(\"hi\")\n```\nworld"
	out := renderMarkdownLite(in, 40, s)
	if !strings.Contains(out, "fmt.Println") {
		t.Fatalf("expected code block content, got %q", out)
	}
}

func TestRenderInlineCode(t *testing.T) {
	s := loadTheme("dark").styles()
	out := renderInlineCode("use `read_file` tool", 40, s)
	if !strings.Contains(out, "read_file") {
		t.Fatalf("expected inline code preserved, got %q", out)
	}
}

func TestRenderToolGroup(t *testing.T) {
	m := Model{
		theme:   loadTheme("dark"),
		styles:  loadTheme("dark").styles(),
		verbose: true,
	}
	out := m.renderToolGroup([]string{"▸ read_file({})", "▸ edit_file({})"}, 60)
	if !strings.Contains(out, "read_file") {
		t.Fatalf("unexpected tool group: %q", out)
	}
}
