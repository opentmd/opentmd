package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m *Model) buildStatusBar() string {
	cfg := m.agent.Config()
	model := cfg.Default.Provider + "/" + cfg.Default.Model
	cwd := collapseHome(m.agent.WorkDir())
	u := m.agent.Usage()

	var left []string
	if model != "" {
		left = append(left, model)
	}
	if cwd != "" {
		left = append(left, cwd)
	}
	if u.TotalTokens > 0 {
		left = append(left, formatTokenCount(u.TotalTokens)+" tok")
	}
	if m.autoRun {
		left = append(left, m.styles.hint.Render("auto-run"))
	}
	leftStr := strings.Join(left, " · ")

	hint := ""
	if m.busy {
		frame := spinnerFrames[m.spinnerFrame%len(spinnerFrames)]
		hint = m.styles.busy.Render(frame + " 流式输出")
	} else if !m.viewport.AtBottom() {
		hint = m.styles.hint.Render("End 回到底部")
	}

	line := m.styles.footer.Render(leftStr)
	if hint != "" {
		gap := m.width - lipgloss.Width(leftStr) - lipgloss.Width(hint) - 2
		if gap < 1 {
			gap = 1
		}
		line = line + strings.Repeat(" ", gap) + hint
	}
	return line
}

func formatTokenCount(n int) string {
	if n >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	}
	if n >= 1_000 {
		return fmt.Sprintf("%.1fk", float64(n)/1_000)
	}
	return fmt.Sprintf("%d", n)
}

func collapseHome(path string) string {
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		if path == home {
			return "~"
		}
		if strings.HasPrefix(path, home+string(os.PathSeparator)) {
			return "~" + path[len(home):]
		}
	}
	return path
}

func (m *Model) buildFooter() string {
	var parts []string
	if m.approvalShowMsg != "" {
		parts = append(parts, m.approvalShowMsg)
	}
	return strings.Join(parts, "\n")
}

func (m *Model) buildInputArea() string {
	width := m.width
	if width < 20 {
		width = 20
	}
	topRule := m.styles.rule.Render(strings.Repeat("─", width))

	var menu string
	if m.mentionMenuVisible() {
		menu = m.renderMentionMenu()
	} else if m.slashMenuVisible() {
		menu = m.renderSlashMenu()
	}

	m.input.SetWidth(width - 4)
	m.input.Prompt = "❯ "
	applyInputStyles(&m.input, m.theme)
	inputView := m.input.View()

	var parts []string
	if menu != "" {
		parts = append(parts, menu)
	}
	parts = append(parts, topRule, inputView)
	botRule := m.styles.rule.Render(strings.Repeat("─", width))
	parts = append(parts, botRule, m.buildStatusBar())
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (m *Model) buildWelcome() string {
	cfg := m.agent.Config()
	model := cfg.Default.Provider + "/" + cfg.Default.Model
	cwd := collapseHome(m.agent.WorkDir())
	w := m.contentWidth()

	brand := m.styles.title.Render("◆ OpenTMD")
	version := m.styles.subtitle.Render(appVersion + "  ·  MIT")

	line1 := lipgloss.JoinHorizontal(lipgloss.Left, brand, strings.Repeat(" ", max(2, w-lipgloss.Width(brand)-lipgloss.Width(version))), version)
	if lipgloss.Width(line1) > w {
		line1 = brand + "\n" + version
	}

	bullet := m.styles.bullet.Render("∙ ")
	pathLine := bullet + m.styles.user.Render(cwd)
	modelLine := bullet + m.styles.user.Render(model)

	hints := m.styles.footer.Render(strings.Join([]string{
		"描述任务，",
		m.styles.hint.Render("/provider"),
		" 配置模型，",
		m.styles.hint.Render("/login"),
		" 登录，",
		m.styles.hint.Render("/help"),
		" 查看命令",
	}, ""))

	return strings.Join([]string{line1, pathLine, modelLine, "", hints}, "\n")
}

func truncatePath(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}
	return "…" + path[len(path)-maxLen+1:]
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
