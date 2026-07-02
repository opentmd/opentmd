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

	// ⚠ BYPASS badge — shown whenever --dangerously-skip-permissions is active.
	if m.agent.AutoApprove() {
		badge := lipgloss.NewStyle().
			Background(m.theme.Warn).
			Foreground(lipgloss.Color("0")).
			Bold(true).
			Render(" " + sym.Warning + " BYPASS ")
		left = append(left, badge)
	}

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
		frame := sym.Spinner[m.spinnerFrame%len(sym.Spinner)]
		hint = m.styles.busy.Render(frame + " 流式输出")
	} else if !m.viewport.AtBottom() {
		hint = m.styles.hint.Render("End 回到底部")
	} else if m.lastAssistantText() != "" {
		hint = m.styles.hint.Render("Ctrl+Y 复制")
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
	topRule := m.styles.rule.Render(strings.Repeat(sym.Rule, width))

	var menu string
	if m.mentionMenuVisible() {
		menu = m.renderMentionMenu()
	} else if m.slashMenuVisible() {
		menu = m.renderSlashMenu()
	}

	m.input.SetWidth(width - 4)
	m.input.Prompt = sym.Prompt
	applyInputStyles(&m.input, m.theme)
	inputView := m.input.View()

	var parts []string
	if menu != "" {
		parts = append(parts, menu)
	}
	parts = append(parts, topRule, inputView)
	botRule := m.styles.rule.Render(strings.Repeat(sym.Rule, width))
	parts = append(parts, botRule, m.buildStatusBar())
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// logoLines is figlet "standard" for "OPEN TMD" — pure ASCII for Windows CMD.
var logoLines = []string{
	`  ___  ____  _____  _   _    _____ __  __ ____`,
	` / _ \|  _ \| ____|| \ | |  |_   _||  \/  ||  _ \`,
	`| | | || |_) |  _|  |  \| |    | | | |\/| || | | |`,
	`| |_| ||  __/| |___ | |\  |    | | | |  | || |_| |`,
	` \___/ |_|   |_____|_| \_|    |_| |_|  |_||____/`,
}

func (m *Model) buildWelcome() string {
	cfg := m.agent.Config()
	model := cfg.Default.Provider + "/" + cfg.Default.Model
	cwd := collapseHome(m.agent.WorkDir())
	w := m.contentWidth()
	h := m.viewport.Height
	if h < 8 {
		h = 8
	}

	var block []string

	const minLogoWidth = 50
	if w >= minLogoWidth {
		maxLogoW := 0
		for _, line := range logoLines {
			if lw := lipgloss.Width(line); lw > maxLogoW {
				maxLogoW = lw
			}
		}
		pad := (w - maxLogoW) / 2
		if pad < 0 {
			pad = 0
		}
		prefix := strings.Repeat(" ", pad)
		for _, line := range logoLines {
			block = append(block, prefix+m.styles.title.Render(line))
		}
		block = append(block, "")
	} else {
		brand := m.styles.title.Render(sym.Diamond + "OpenTMD")
		block = append(block, brand, "")
	}

	version := m.styles.subtitle.Render(appVersion + "  ·  MIT")
	block = append(block, lipgloss.Place(w, lipgloss.Height(version), lipgloss.Center, lipgloss.Top, version))

	bullet := m.styles.bullet.Render(sym.Bullet)
	pathLine := bullet + m.styles.user.Render(cwd)
	modelLine := bullet + m.styles.user.Render(model)
	block = append(block, "", pathLine, modelLine, "")

	hints := m.styles.footer.Render(strings.Join([]string{
		"描述任务，",
		m.styles.hint.Render("/provider"),
		" 配置模型，",
		m.styles.hint.Render("/login"),
		" 登录，",
		m.styles.hint.Render("/help"),
		" 查看命令",
	}, ""))
	block = append(block, hints)

	content := strings.Join(block, "\n")
	return lipgloss.Place(
		w,
		h,
		lipgloss.Center,
		lipgloss.Center,
		content,
		lipgloss.WithWhitespaceChars(" "),
	)
}

func truncatePath(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}
	return sym.Ellipsis + path[len(path)-maxLen+len(sym.Ellipsis):]
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
