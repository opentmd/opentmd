package tui

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/mattn/go-runewidth"

	"github.com/opentmd/opentmd-cli/internal/permission"
)

type displayMsg struct {
	kind string // user, assistant, system, tool
	text string
}

func wrapPlainText(text string, width int) string {
	if width < 10 {
		width = 10
	}
	var lines []string
	for _, paragraph := range strings.Split(text, "\n") {
		if paragraph == "" {
			lines = append(lines, "")
			continue
		}
		lines = append(lines, wrapParagraph(paragraph, width)...)
	}
	return strings.Join(lines, "\n")
}

func wrapParagraph(text string, width int) []string {
	words := splitWords(text)
	if len(words) == 0 {
		return []string{""}
	}

	var lines []string
	var current strings.Builder
	currentWidth := 0

	flush := func() {
		if current.Len() > 0 {
			lines = append(lines, current.String())
			current.Reset()
			currentWidth = 0
		}
	}

	for _, word := range words {
		ww := runewidth.StringWidth(word)
		if ww > width {
			flush()
			lines = append(lines, splitLongWord(word, width)...)
			continue
		}
		extra := ww
		if currentWidth > 0 {
			extra = ww + 1
		}
		if currentWidth+extra > width {
			flush()
		}
		if currentWidth > 0 {
			current.WriteByte(' ')
			currentWidth++
		}
		current.WriteString(word)
		currentWidth += ww
	}
	flush()
	return lines
}

func splitWords(s string) []string {
	var words []string
	var b strings.Builder
	for _, r := range s {
		if unicode.IsSpace(r) {
			if b.Len() > 0 {
				words = append(words, b.String())
				b.Reset()
			}
		} else {
			b.WriteRune(r)
		}
	}
	if b.Len() > 0 {
		words = append(words, b.String())
	}
	return words
}

func splitLongWord(word string, width int) []string {
	var parts []string
	for word != "" {
		var b strings.Builder
		w := 0
		for _, r := range word {
			rw := runewidth.RuneWidth(r)
			if w+rw > width {
				break
			}
			b.WriteRune(r)
			w += rw
		}
		chunk := b.String()
		if chunk == "" {
			r, size := utf8.DecodeRuneInString(word)
			if size == 0 {
				break
			}
			chunk = string(r)
		}
		parts = append(parts, chunk)
		word = strings.TrimPrefix(word, chunk)
	}
	return parts
}

func (m *Model) showApproval(req permission.Request) {
	m.approvalPending = &req
	title := "⚠ 需要权限审批"
	if req.Reason != "" {
		title = "⚠ " + req.Reason
	}
	body := fmt.Sprintf(
		"工具: %s\n参数: %s\n\n[y/Y] 允许一次   [a/A] 本会话允许   [n/N] 拒绝",
		req.Tool,
		truncate(req.Args, 160),
	)
	box := lipgloss.NewStyle().
		Foreground(m.theme.Warn).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Warn).
		Padding(0, 1).
		Width(min(m.width-6, 72)).
		Render(title + "\n" + body)
	m.approvalShowMsg = box
}

func (m *Model) handleApprovalKey(key string) bool {
	if m.approvalPending == nil || m.approvalCh == nil {
		return false
	}
	switch strings.ToLower(key) {
	case "y":
		m.approvalCh <- permission.AllowOnce
		m.appendSystem("✓ 已允许（一次）")
	case "a":
		m.approvalCh <- permission.AllowSession
		m.appendSystem("✓ 已允许（本会话）")
	case "n":
		m.approvalCh <- permission.Deny
		m.appendSystem("✗ 已拒绝")
	default:
		return false
	}
	m.approvalPending = nil
	m.approvalShowMsg = ""
	m.refreshViewport()
	return true
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func (m *Model) renderMessages(contentWidth int) string {
	if len(m.messages) == 0 {
		return m.buildWelcome()
	}
	if contentWidth < 10 {
		contentWidth = 10
	}
	var sb strings.Builder
	lastAI := m.lastAssistantIdx()
	for i := 0; i < len(m.messages); i++ {
		msg := m.messages[i]
		if msg.kind == "tool" {
			var tools []string
			for i < len(m.messages) && m.messages[i].kind == "tool" {
				tools = append(tools, strings.TrimSpace(m.messages[i].text))
				i++
			}
			i--
			sb.WriteString(m.renderToolGroup(tools, contentWidth))
		} else {
			if sb.Len() > 0 {
				sb.WriteByte('\n')
			}
			sb.WriteString(m.renderOneMessage(msg, contentWidth, i, lastAI))
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}

func (m *Model) renderUserEcho(text string, width int) string {
	chev := m.styles.userChevron.Render("❯ ")
	body := wrapPlainText(text, width-2)
	var lines []string
	for i, line := range strings.Split(body, "\n") {
		if i == 0 {
			lines = append(lines, chev+m.styles.user.Render(line))
		} else {
			lines = append(lines, "  "+m.styles.user.Render(line))
		}
	}
	return strings.Join(lines, "\n")
}

func (m *Model) renderAssistant(text string, width int, markdown bool) string {
	wrapped := text
	if markdown {
		wrapped = renderMarkdownLite(text, width, m.styles)
	} else {
		wrapped = wrapPlainText(text, width)
	}
	return m.styles.assistant.Render(wrapped)
}

func (m *Model) renderToolGroup(lines []string, contentWidth int) string {
	if len(lines) == 0 {
		return ""
	}
	var cleaned []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		cleaned = append(cleaned, line)
	}
	if len(cleaned) == 0 {
		return ""
	}
	if !m.verbose {
		return m.styles.toolHeader.Render(fmt.Sprintf("● %d 个工具步骤", len(cleaned)))
	}
	var out []string
	for _, line := range cleaned {
		line = strings.TrimPrefix(line, "▸ ")
		parts := strings.SplitN(line, " ", 2)
		name := parts[0]
		detail := ""
		if len(parts) > 1 {
			detail = parts[1]
		}
		row := m.styles.toolBullet.Render("● ") + lipgloss.NewStyle().Bold(true).Render(name)
		if detail != "" {
			row += " " + m.styles.toolHeader.Render(detail)
		}
		out = append(out, row)
	}
	return strings.Join(out, "\n")
}

func (m *Model) renderOneMessage(msg displayMsg, contentWidth, idx, lastAI int) string {
	switch msg.kind {
	case "user":
		return m.renderUserEcho(msg.text, contentWidth)
	case "assistant":
		if strings.TrimSpace(msg.text) == "" && m.busy {
			frame := spinnerFrames[m.spinnerFrame%len(spinnerFrames)]
			return m.styles.footer.Render("  " + frame + " …")
		}
		text := msg.text
		streaming := m.busy && idx == lastAI
		if streaming && text != "" {
			text += "▌"
		}
		return m.renderAssistant(text, contentWidth, !streaming)
	case "tool":
		line := "▸ " + msg.text
		wrapped := wrapPlainText(line, contentWidth-4)
		return m.styles.tool.Width(contentWidth - 2).Render(wrapped)
	default:
		return m.renderSystem(msg.text, contentWidth)
	}
}

func (m *Model) renderSystem(text string, contentWidth int) string {
	wrapped := wrapPlainText(text, contentWidth)
	switch {
	case strings.HasPrefix(text, "✓"):
		return m.styles.success.Render(wrapped)
	case strings.HasPrefix(text, "✗"):
		return m.styles.error.Render(wrapped)
	case strings.HasPrefix(text, "⏹"):
		return m.styles.warn.Render(wrapped)
	default:
		return m.styles.system.Render(wrapped)
	}
}

func (m *Model) renderBubble(label, text string, width int, style lipgloss.Style, markdown bool) string {
	labelTag := m.styles.subtitle.Render(" " + label + " ")
	prefixWidth := runewidth.StringWidth(ansi.Strip(labelTag)) + 1
	bodyWidth := width - prefixWidth - 2
	if bodyWidth < 10 {
		bodyWidth = 10
	}
	var wrapped string
	if markdown {
		wrapped = renderMarkdownLite(text, bodyWidth, m.styles)
	} else {
		wrapped = wrapPlainText(text, bodyWidth)
	}
	lines := strings.Split(wrapped, "\n")
	var out strings.Builder
	for i, line := range lines {
		if i == 0 {
			out.WriteString(style.Render(labelTag + line))
		} else {
			out.WriteString(style.Render(strings.Repeat(" ", prefixWidth) + line))
		}
		if i < len(lines)-1 {
			out.WriteByte('\n')
		}
	}
	return out.String()
}

func (m *Model) appendMsg(kind, text string) {
	m.messages = append(m.messages, displayMsg{kind: kind, text: text})
}

func (m *Model) lastAssistantIdx() int {
	for i := len(m.messages) - 1; i >= 0; i-- {
		if m.messages[i].kind == "assistant" {
			return i
		}
	}
	return -1
}

func (m *Model) setAssistantContent(content string) {
	if idx := m.lastAssistantIdx(); idx >= 0 {
		if strings.TrimSpace(m.messages[idx].text) == "" {
			m.messages[idx].text = content
		}
		return
	}
	m.appendMsg("assistant", content)
}

func (m *Model) appendSystem(text string) {
	m.appendMsg("system", text)
}

func (m *Model) contentWidth() int {
	w := m.viewport.Width - m.viewport.Style.GetHorizontalFrameSize()
	if w < 10 {
		w = 10
	}
	return w
}
