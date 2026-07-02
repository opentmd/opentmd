package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/opentmd/opentmd/internal/trust"
)

func (m *Model) initTrust(opts trust.Options) {
	if opts.Skip {
		return
	}
	if opts.AutoTrust {
		_ = trust.Trust(m.agent.WorkDir())
		return
	}
	ok, err := trust.IsTrusted(m.agent.WorkDir())
	if err != nil || ok {
		return
	}
	m.trustPending = true
	m.trustWorkDir = m.agent.WorkDir()
	m.trustSelected = 0
	m.showTrustPrompt()
}

func (m *Model) requireTrustForDir(workDir string) {
	ok, err := trust.IsTrusted(workDir)
	if err != nil || ok {
		return
	}
	m.trustPending = true
	m.trustWorkDir = workDir
	m.trustSelected = 0
	m.showTrustPrompt()
}

func (m *Model) showTrustPrompt() {
	m.trustShowMsg = m.renderTrustBox(m.trustWorkDir, m.trustSelected)
}

func (m *Model) renderTrustBox(workDir string, selected int) string {
	width := min(m.width-6, 72)
	if width < 40 {
		width = 40
	}

	title := "⚠ Workspace Trust Required"
	body := strings.Join([]string{
		"",
		"OpenTMD can execute code and access files in this directory.",
		"",
		"Do you trust the contents of this directory?",
		"",
		workDir,
		"",
	}, "\n")

	trustLine := "  ▶ [a] Trust this workspace"
	quitLine := "    [q] Quit"
	if selected == 1 {
		trustLine = "    [a] Trust this workspace"
		quitLine = "  ▶ [q] Quit"
	}
	footer := strings.Join([]string{
		trustLine,
		quitLine,
		"",
		"Use arrow keys to navigate, Enter to select, or press the key shown",
	}, "\n")

	content := title + "\n" + body + footer
	return lipgloss.NewStyle().
		Foreground(m.theme.Warn).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Warn).
		Padding(1, 2).
		Width(width).
		Render(content)
}

func (m *Model) renderTrustScreen() string {
	if !m.ready {
		return m.styles.subtitle.Render("\n  初始化 OpenTMD…\n")
	}
	box := m.trustShowMsg
	if box == "" {
		box = m.renderTrustBox(m.trustWorkDir, m.trustSelected)
	}
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		box,
		lipgloss.WithWhitespaceChars(" "),
	)
}

func (m *Model) handleTrustKey(key string) bool {
	if !m.trustPending {
		return false
	}
	switch key {
	case "up", "k":
		if m.trustSelected > 0 {
			m.trustSelected--
			m.showTrustPrompt()
		}
		return true
	case "down", "j":
		if m.trustSelected < 1 {
			m.trustSelected++
			m.showTrustPrompt()
		}
		return true
	case "enter", "a":
		if m.trustSelected == 0 {
			if err := trust.Trust(m.trustWorkDir); err != nil {
				m.trustShowMsg = m.renderTrustBox(m.trustWorkDir, m.trustSelected) +
					"\n" + fmt.Sprintf("✗ 无法保存信任: %v", err)
				return true
			}
			m.trustPending = false
			m.trustShowMsg = ""
			m.appendSystem("✓ 已信任工作区: " + m.trustWorkDir)
			m.refreshViewport()
			return true
		}
		return m.quitFromTrust()
	case "q":
		return m.quitFromTrust()
	default:
		return false
	}
}

func (m *Model) quitFromTrust() bool {
	m.trustPending = false
	m.trustShowMsg = ""
	if m.program != nil {
		m.program.Quit()
	}
	return true
}
