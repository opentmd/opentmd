package tui

import (
	"context"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textarea"
)

const doubleCtrlCWindow = 2 * time.Second

func keybindingsHelp() string {
	return strings.Join([]string{
		"输入:",
		"  Enter              发送",
		"  Ctrl+J / Alt+Enter 换行",
		"  \\ + Enter          换行（行尾反斜杠）",
		"  Tab                斜杠命令补全",
		"  Esc                取消输出 / 清空输入",
		"  ↑ / ↓              输入历史",
		"  Ctrl+U             清空整行",
		"",
		"滚动:",
		"  Shift+↑↓           逐行滚动",
		"  PgUp / PgDn        翻页",
		"  Alt+↑↓             上/下一条消息",
		"  Ctrl+↑↓            上/下一条用户消息",
		"  Home / End         顶/底",
		"  Ctrl+G             滚到最新",
		"",
		"复制:",
		"  Ctrl+Y             复制最近 AI 输出到剪贴板",
		"  /copy              同上；/copy all 复制整个会话",
		"  Shift+拖拽         终端原生文字选中",
		"",
		"其他:",
		"  Ctrl+O             切换工具详情（verbose）",
		"  Ctrl+L             清屏（保留对话）",
		"  Ctrl+C             取消；连按两次退出",
		"  Ctrl+D             空输入时退出",
		"  cd /path           切换目录（无需 /）",
	}, "\n")
}

func lastInputLine(val string) string {
	if idx := strings.LastIndex(val, "\n"); idx >= 0 {
		return val[idx+1:]
	}
	return val
}

func endsWithContinuation(val string) bool {
	return strings.HasSuffix(strings.TrimRight(lastInputLine(val), " \t"), "\\")
}

func stripContinuationLine(line string) string {
	line = strings.TrimRight(line, " \t")
	if strings.HasSuffix(line, "\\") {
		return strings.TrimRight(line[:len(line)-1], " \t")
	}
	return line
}

func (m *Model) shouldInsertNewline(msg tea.KeyMsg) bool {
	if msg.Alt {
		return true
	}
	switch msg.Type {
	case tea.KeyCtrlJ:
		return true
	case tea.KeyEnter:
		return endsWithContinuation(m.input.Value())
	}
	return false
}

func (m *Model) insertNewline() tea.Cmd {
	val := m.input.Value()
	if endsWithContinuation(val) {
		lines := strings.Split(val, "\n")
		last := lines[len(lines)-1]
		lines[len(lines)-1] = stripContinuationLine(last)
		val = strings.Join(lines, "\n")
	}
	m.input.SetValue(val + "\n")
	return textarea.Blink
}

func (m *Model) handleQuitKey() (tea.Model, tea.Cmd) {
	if m.busy && m.chatCancel != nil {
		m.chatCancel()
		m.appendSystem("⏹ 已取消")
		m.refreshViewport()
		return m, nil
	}
	now := time.Now()
	if !m.lastCtrlC.IsZero() && now.Sub(m.lastCtrlC) < doubleCtrlCWindow {
		return m, tea.Quit
	}
	m.lastCtrlC = now
	return m, nil
}

func (m *Model) handleEscKey() (tea.Model, tea.Cmd) {
	if m.busy && m.chatCancel != nil {
		m.chatCancel()
		m.appendSystem("⏹ 已取消")
		m.refreshViewport()
		return m, nil
	}
	m.input.Reset()
	m.historyIndex = -1
	m.historyDraft = ""
	return m, nil
}

func (m *Model) handleCtrlD() (tea.Model, tea.Cmd) {
	if strings.TrimSpace(m.input.Value()) != "" {
		return m, nil
	}
	return m, tea.Quit
}

func (m *Model) handleHistoryKey(key tea.KeyType) bool {
	if m.inputHistory == nil {
		return false
	}
	switch key {
	case tea.KeyUp:
		entry, idx := m.inputHistory.prev(m.historyDraft, m.historyIndex)
		if idx < 0 {
			return true
		}
		if m.historyIndex < 0 {
			m.historyDraft = m.input.Value()
		}
		m.historyIndex = idx
		m.input.SetValue(entry)
		return true
	case tea.KeyDown:
		if m.historyIndex < 0 {
			return true
		}
		entry, idx := m.inputHistory.next(m.historyIndex)
		if idx < 0 {
			m.input.SetValue(m.historyDraft)
			m.historyIndex = -1
			m.historyDraft = ""
		} else {
			m.historyIndex = idx
			m.input.SetValue(entry)
		}
		return true
	}
	return false
}

func (m *Model) submitInput() (tea.Model, tea.Cmd) {
	if m.busy {
		return m, nil
	}
	input := strings.TrimSpace(m.input.Value())
	if input == "" {
		return m, nil
	}

	if m.inputHistory != nil {
		_ = m.inputHistory.add(input)
	}
	m.historyIndex = -1
	m.historyDraft = ""
	m.input.Reset()

	if strings.HasPrefix(input, "cd ") {
		dir := strings.TrimSpace(strings.TrimPrefix(input, "cd "))
		if dir != "" {
			return m.handleBareCd(dir)
		}
	}

	if strings.HasPrefix(input, "/") {
		return m.handleSlash(input)
	}

	m.appendMsg("user", input)
	m.appendMsg("assistant", "")
	m.followOutput = true
	m.refreshViewport()
	m.busy = true
	ctx, cancel := context.WithCancel(context.Background())
	m.chatCancel = cancel
	return m, m.streamChat(ctx, input)
}

func (m *Model) handleBareCd(dir string) (tea.Model, tea.Cmd) {
	if err := m.agent.SetWorkDir(dir); err != nil {
		m.appendSystem("✗ 切换目录失败: " + err.Error())
	} else {
		m.appendSystem("✓ 工作目录: " + m.agent.WorkDir())
		m.requireTrustForDir(m.agent.WorkDir())
		m.invalidateFileIndex()
	}
	m.refreshViewport()
	return m, nil
}
