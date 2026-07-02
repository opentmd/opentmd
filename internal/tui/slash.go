package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

const slashMenuMaxVisible = 4

type slashEntry struct {
	Name string
	Args string
	Desc string
}

func (e slashEntry) Label() string {
	if e.Args == "" {
		return e.Name
	}
	return e.Name + " " + e.Args
}

var slashCommands = []slashEntry{
	{Name: "/help", Desc: "显示全部命令"},
	{Name: "/keys", Desc: "显示快捷键"},
	{Name: "/status", Desc: "当前 provider / 目录 / session"},
	{Name: "/config", Desc: "配置文件路径与摘要"},
	{Name: "/cost", Desc: "Token 用量与估算费用"},
	{Name: "/resume", Desc: "恢复历史 session"},
	{Name: "/session", Desc: "开始新会话"},
	{Name: "/bg", Args: "list", Desc: "后台 session 管理"},
	{Name: "/clear", Desc: "清屏（保留对话）"},
	{Name: "/plan", Desc: "切换到只读探索模式"},
	{Name: "/build", Desc: "切换到可执行构建模式"},
	{Name: "/undo", Desc: "撤销上一轮文件修改"},
	{Name: "/auto-run", Args: "on", Desc: "自动执行工具（跳过确认）"},
	{Name: "/model", Desc: "切换 provider / model"},
	{Name: "/provider", Desc: "Provider 管理向导"},
	{Name: "/login", Desc: "配置 Provider 与 API Key"},
	{Name: "/cd", Args: "<dir>", Desc: "切换工作目录"},
	{Name: "/reload", Desc: "重载 config.toml"},
	{Name: "/lsp", Desc: "LSP 语言服务器状态"},
	{Name: "/skills", Desc: "列出 Skills"},
	{Name: "/mcp", Desc: "MCP 服务器状态"},
	{Name: "/mcp", Args: "reload", Desc: "重连 MCP 服务器"},
	{Name: "/copy", Args: "all", Desc: "复制输出到剪贴板"},
	{Name: "/export", Args: "[file]", Desc: "导出会话到文件"},
	{Name: "/remember", Args: "<text>", Desc: "记忆一条全局事实"},
	{Name: "/remember", Args: "project <text>", Desc: "记忆一条项目事实"},
	{Name: "/forget", Args: "<keyword>", Desc: "删除匹配的记忆"},
	{Name: "/memory", Desc: "查看所有记忆"},
	{Name: "/compact", Args: "[focus]", Desc: "压缩对话历史"},
	{Name: "/quit", Desc: "退出程序"},
	{Name: "/exit", Desc: "同 /quit"},
}

func filterSlashCommands(query string) []slashEntry {
	query = strings.TrimSpace(query)
	if query == "" || query == "/" {
		out := make([]slashEntry, len(slashCommands))
		copy(out, slashCommands)
		return out
	}
	var out []slashEntry
	for _, c := range slashCommands {
		if strings.HasPrefix(c.Name, query) {
			out = append(out, c)
		}
	}
	return out
}

func clampSlashSelection(selected, count int) int {
	if count <= 0 {
		return 0
	}
	if selected < 0 {
		return 0
	}
	if selected >= count {
		return count - 1
	}
	return selected
}

func slashMenuWindow(selected, total, maxVisible, scroll int) (start, end, nextScroll int) {
	if maxVisible <= 0 {
		maxVisible = slashMenuMaxVisible
	}
	if total <= maxVisible {
		return 0, total, 0
	}
	nextScroll = scroll
	if selected < nextScroll {
		nextScroll = selected
	}
	if selected >= nextScroll+maxVisible {
		nextScroll = selected - maxVisible + 1
	}
	start = nextScroll
	end = start + maxVisible
	if end > total {
		end = total
		start = end - maxVisible
		nextScroll = start
	}
	return start, end, nextScroll
}

func (m *Model) slashFirstLine() string {
	val := m.input.Value()
	if idx := strings.Index(val, "\n"); idx >= 0 {
		return val[:idx]
	}
	return val
}

func (m *Model) slashQuery() string {
	line := strings.TrimSpace(m.slashFirstLine())
	if !strings.HasPrefix(line, "/") {
		return ""
	}
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return "/"
	}
	return parts[0]
}

func (m *Model) slashMenuVisible() bool {
	if m.trustPending || m.approvalPending != nil || m.modalActive() {
		return false
	}
	line := strings.TrimSpace(m.slashFirstLine())
	return strings.HasPrefix(line, "/")
}

func (m *Model) filteredSlashCommands() []slashEntry {
	return filterSlashCommands(m.slashQuery())
}

func (m *Model) slashMenuLineCount() int {
	if !m.slashMenuVisible() {
		return 0
	}
	items := m.filteredSlashCommands()
	if len(items) == 0 {
		return 2
	}
	visible := len(items)
	if visible > slashMenuMaxVisible {
		visible = slashMenuMaxVisible
	}
	lines := visible + 1
	if len(items) > slashMenuMaxVisible {
		lines++
	}
	return lines
}

func (m *Model) syncSlashMenu() {
	if !m.slashMenuVisible() {
		m.slashLastQuery = ""
		m.slashSelected = 0
		m.slashScroll = 0
		return
	}
	query := m.slashQuery()
	if query != m.slashLastQuery {
		m.slashSelected = 0
		m.slashScroll = 0
		m.slashLastQuery = query
	}
	items := m.filteredSlashCommands()
	m.slashSelected = clampSlashSelection(m.slashSelected, len(items))
}

func (m *Model) handleSlashMenuKey(msg tea.KeyMsg) (bool, tea.Model, tea.Cmd) {
	if !m.slashMenuVisible() {
		return false, m, nil
	}
	items := m.filteredSlashCommands()
	if len(items) == 0 {
		return false, m, nil
	}

	switch msg.Type {
	case tea.KeyUp, tea.KeyDown, tea.KeyCtrlP, tea.KeyCtrlN:
		m.moveSlashSelection(msg.Type)
		return true, m, nil
	case tea.KeyTab:
		m.applySlashCompletion()
		return true, m, nil
	case tea.KeyEnter:
		if msg.Alt {
			return false, m, nil
		}
		input := strings.TrimSpace(m.input.Value())
		if input == "" {
			return false, m, nil
		}
		if m.shouldUseSlashSelection(input, items) {
			sel := items[clampSlashSelection(m.slashSelected, len(items))]
			input = sel.Name
			if len(strings.Fields(strings.TrimSpace(m.slashFirstLine()))) > 1 {
				input = strings.TrimSpace(m.slashFirstLine())
			}
		}
		m.input.Reset()
		model, cmd := m.handleSlash(input)
		return true, model, cmd
	}
	return false, m, nil
}

func cycleSlashSelection(selected, count int, up bool) int {
	if count <= 0 {
		return 0
	}
	if up {
		if selected <= 0 {
			return count - 1
		}
		return selected - 1
	}
	if selected >= count-1 {
		return 0
	}
	return selected + 1
}

func (m *Model) moveSlashSelection(key tea.KeyType) {
	items := m.filteredSlashCommands()
	if len(items) == 0 {
		return
	}
	switch key {
	case tea.KeyUp, tea.KeyCtrlP:
		m.slashSelected = cycleSlashSelection(m.slashSelected, len(items), true)
	case tea.KeyDown, tea.KeyCtrlN:
		m.slashSelected = cycleSlashSelection(m.slashSelected, len(items), false)
	}
	_, _, m.slashScroll = slashMenuWindow(m.slashSelected, len(items), slashMenuMaxVisible, m.slashScroll)
}

func (m *Model) shouldUseSlashSelection(input string, items []slashEntry) bool {
	first := strings.TrimSpace(m.slashFirstLine())
	parts := strings.Fields(first)
	if len(parts) != 1 {
		return false
	}
	cmd := parts[0]
	for _, item := range items {
		if item.Name == cmd {
			return false
		}
	}
	return true
}

func (m *Model) applySlashCompletion() {
	items := m.filteredSlashCommands()
	if len(items) == 0 {
		return
	}
	sel := items[clampSlashSelection(m.slashSelected, len(items))]
	value := sel.Name
	if sel.Args != "" {
		value += " "
	}
	rest := ""
	if val := m.input.Value(); strings.Contains(val, "\n") {
		if idx := strings.Index(val, "\n"); idx >= 0 {
			rest = val[idx:]
		}
	}
	m.input.SetValue(value + rest)
	m.slashLastQuery = sel.Name
}

func (m *Model) renderSlashMenu() string {
	if !m.slashMenuVisible() {
		return ""
	}
	items := m.filteredSlashCommands()
	if len(items) == 0 {
		return m.styles.hint.Render("  无匹配命令 · Tab 补全 · ↑↓ 选择")
	}

	selected := clampSlashSelection(m.slashSelected, len(items))
	start, end, scroll := slashMenuWindow(selected, len(items), slashMenuMaxVisible, m.slashScroll)
	m.slashScroll = scroll

	labelWidth := 0
	for i := start; i < end; i++ {
		if w := runewidth.StringWidth(items[i].Label()); w > labelWidth {
			labelWidth = w
		}
	}
	if labelWidth < 12 {
		labelWidth = 12
	}

	activeStyle := m.styles.menuActive
	inactiveStyle := m.styles.menuInactive
	descActive := m.styles.menuDesc
	descInactive := m.styles.menuDesc

	var lines []string
	for i := start; i < end; i++ {
		item := items[i]
		prefix := "▸ "
		nameStyle := inactiveStyle
		descStyle := descInactive
		if i == selected {
			nameStyle = activeStyle
			descStyle = descActive
		}
		label := nameStyle.Render(prefix + item.Label())
		padding := labelWidth - runewidth.StringWidth(item.Label())
		if padding < 1 {
			padding = 1
		}
		desc := descStyle.Render(item.Desc)
		lines = append(lines, prefix+label+strings.Repeat(" ", padding)+"  "+desc)
	}

	if len(items) > slashMenuMaxVisible {
		if end < len(items) {
			lines = append(lines, m.styles.hint.Render("  ↓ more below"))
		} else if start > 0 {
			lines = append(lines, m.styles.hint.Render("  ↑ more above"))
		}
	}

	hint := m.styles.inputBar.Render("Tab 补全 · Enter 执行 · ↑↓ 循环")
	return lipgloss.JoinVertical(lipgloss.Left, strings.Join(lines, "\n"), hint)
}
