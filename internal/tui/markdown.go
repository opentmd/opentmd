package tui

import (
	"strings"
)

// renderMarkdownLite applies lightweight formatting: fenced code blocks and `inline code`.
func renderMarkdownLite(text string, width int, s themeStyles) string {
	if !strings.Contains(text, "```") && !strings.Contains(text, "`") {
		return wrapPlainText(text, width)
	}

	parts := strings.Split(text, "```")
	var sb strings.Builder
	for i, part := range parts {
		if i%2 == 1 {
			code := strings.TrimSpace(stripCodeFenceLang(part))
			if code == "" {
				continue
			}
			wrapped := wrapPlainText(code, max(10, width-4))
			if sb.Len() > 0 {
				sb.WriteByte('\n')
			}
			sb.WriteString(s.code.Render(wrapped))
		} else {
			if part != "" {
				if sb.Len() > 0 {
					sb.WriteByte('\n')
				}
				sb.WriteString(renderInlineCode(part, width, s))
			}
		}
	}
	return sb.String()
}

func stripCodeFenceLang(block string) string {
	block = strings.TrimLeft(block, "\n")
	if idx := strings.Index(block, "\n"); idx >= 0 {
		first := strings.TrimSpace(block[:idx])
		if first != "" && !strings.Contains(first, " ") && len(first) <= 24 {
			return block[idx+1:]
		}
	}
	return block
}

func renderInlineCode(text string, width int, s themeStyles) string {
	if !strings.Contains(text, "`") {
		return wrapPlainText(text, width)
	}
	parts := strings.Split(text, "`")
	var sb strings.Builder
	for i, part := range parts {
		if i%2 == 1 {
			sb.WriteString(s.codeInline.Render(part))
		} else {
			sb.WriteString(part)
		}
	}
	return wrapPlainText(sb.String(), width)
}
