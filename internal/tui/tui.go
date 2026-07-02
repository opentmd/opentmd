package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/opentmd/opentmd/internal/agent"
	"github.com/opentmd/opentmd/internal/config"
	"github.com/opentmd/opentmd/internal/filehistory"
	"github.com/opentmd/opentmd/internal/llm"
	"github.com/opentmd/opentmd/internal/lsp"
	"github.com/opentmd/opentmd/internal/permission"
	"github.com/opentmd/opentmd/internal/pricing"
	"github.com/opentmd/opentmd/internal/trust"
)

type chunkMsg struct {
	content string
}

type statusMsg struct {
	text string
}

type doneMsg struct {
	err     error
	content string
}

type compactDoneMsg struct {
	text string
	err  error
}

type approvalMsg struct {
	req permission.Request
}

type tickMsg struct{}

type lspConnectMsg struct {
	ev lsp.ConnectEvent
}

type Model struct {
	agent        *agent.Agent
	program      *tea.Program
	viewport     viewport.Model
	input        textarea.Model
	messages     []displayMsg
	width        int
	height       int
	ready        bool
	busy         bool
	followOutput bool
	chatCancel   context.CancelFunc
	spinnerFrame int

	theme  Theme
	styles themeStyles

	approvalCh      chan permission.Decision
	approvalPending *permission.Request
	approvalShowMsg string

	trustPending  bool
	trustWorkDir  string
	trustSelected int
	trustShowMsg  string
	trustOpts     trust.Options

	slashSelected  int
	slashScroll    int
	slashLastQuery string

	mentionSelected  int
	mentionScroll    int
	mentionLastQuery string
	fileIndex        *fileIndex

	bgSessions []bgSession

	modalKind       modalKind
	modalItems      []pickerItem
	modalSelected   int
	modalTextBuf    string
	modalTextAction string
	modalTextTarget string
	modalTextHint   string
	modalTextMask   bool
	loginPreset     string
	loginAPIKey     string

	inputHistory   *inputHistory
	historyIndex   int
	historyDraft   string
	lastCtrlC      time.Time
	messageAnchors []messageAnchor
	verbose        bool
	autoRun        bool
}

func applyInputStyles(ta *textarea.Model, theme Theme) {
	is := theme.inputStyles()
	ta.FocusedStyle.Prompt = is.focused.prompt
	ta.FocusedStyle.Text = is.focused.text
	ta.FocusedStyle.Placeholder = is.focused.placeholder
	ta.FocusedStyle.CursorLine = is.focused.cursorLine
	ta.BlurredStyle.Prompt = is.blurred.prompt
	ta.BlurredStyle.Text = is.blurred.text
	ta.BlurredStyle.Placeholder = is.blurred.placeholder
}

func New(ag *agent.Agent, opts trust.Options) Model {
	theme := loadTheme(ag.Config().TUI.Theme)
	styles := theme.styles()

	ta := textarea.New()
	ta.Placeholder = "描述任务…  (/help  /provider  /login)"
	ta.Focus()
	ta.CharLimit = 4096
	ta.SetWidth(60)
	ta.SetHeight(1)
	ta.ShowLineNumbers = false
	ta.Prompt = "❯ "
	applyInputStyles(&ta, theme)

	vp := viewport.New(80, 20)
	vp.Style = styles.viewport
	vp.MouseWheelEnabled = true
	vp.SetContent("")

	m := Model{
		agent:        ag,
		viewport:     vp,
		input:        ta,
		followOutput: true,
		theme:        theme,
		styles:       styles,
		trustOpts:    opts,
	}
	m.initTrust(opts)
	if h, err := loadInputHistory(); err == nil {
		m.inputHistory = h
	}
	m.loadSessionMessages()
	return m
}

func (m *Model) loadSessionMessages() {
	m.messages = nil
	sess := m.agent.Session().Current()
	if sess == nil {
		m.refreshViewport()
		return
	}
	for _, msg := range sess.Messages {
		switch msg.Role {
		case llm.RoleUser:
			m.appendMsg("user", msg.Content)
		case llm.RoleAssistant:
			m.appendMsg("assistant", msg.Content)
		}
	}
	m.refreshViewport()
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, m.tick())
}

func (m *Model) tick() tea.Cmd {
	return tea.Tick(120*time.Millisecond, func(time.Time) tea.Msg { return tickMsg{} })
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.adjustLayout()
		m.refreshViewport()
		return m, nil

	case tickMsg:
		if m.busy {
			m.spinnerFrame++
			m.refreshViewport()
		}
		return m, m.tick()

	case chunkMsg:
		if idx := m.lastAssistantIdx(); idx >= 0 {
			m.messages[idx].text += msg.content
		}
		m.followOutput = true
		m.refreshViewport()
		return m, nil

	case statusMsg:
		if n := len(m.messages); n > 0 && m.messages[n-1].kind == "tool" {
			m.messages[n-1].text += "\n" + msg.text
		} else {
			m.appendMsg("tool", msg.text)
		}
		m.followOutput = true
		m.refreshViewport()
		return m, nil

	case approvalMsg:
		m.showApproval(msg.req)
		return m, nil

	case doneMsg:
		m.busy = false
		m.chatCancel = nil
		if msg.err != nil && msg.err != context.Canceled {
			m.appendSystem("✗ 错误: " + msg.err.Error())
		} else if msg.content != "" {
			m.setAssistantContent(msg.content)
		}
		m.refreshViewport()
		return m, nil

	case compactDoneMsg:
		m.busy = false
		m.chatCancel = nil
		if msg.err != nil && msg.err != context.Canceled {
			m.appendSystem("✗ 压缩失败: " + msg.err.Error())
		} else if msg.text != "" {
			m.loadSessionMessages()
			m.appendSystem(msg.text)
		}
		m.refreshViewport()
		return m, nil

	case lspConnectMsg:
		if msg.ev.Started {
			m.appendSystem(fmt.Sprintf("✓ LSP %s started (.%s)", msg.ev.Command, msg.ev.Ext))
		} else if msg.ev.Error != "" {
			m.appendSystem(fmt.Sprintf("× LSP %s failed (.%s): %s", msg.ev.Command, msg.ev.Ext, msg.ev.Error))
		}
		m.refreshViewport()
		return m, nil

	case tea.QuitMsg:
		return m, tea.Quit

	case tea.MouseMsg:
		if m.handleViewportScroll(msg) {
			return m, nil
		}

	case tea.KeyMsg:
		if m.trustPending {
			if m.handleTrustKey(msg.String()) {
				return m, nil
			}
		}
		if m.approvalPending != nil {
			if m.handleApprovalKey(msg.String()) {
				return m, nil
			}
		}
		if m.handleModalKey(msg) {
			m.refreshViewport()
			return m, nil
		}
		if isQuitKey(msg) {
			return m.handleQuitKey()
		}
		if isClearInputKey(msg) && !m.modalActive() {
			return m.handleEscKey()
		}
		if msg.Type == tea.KeyCtrlD {
			if model, cmd := m.handleCtrlD(); cmd != nil {
				return model, cmd
			}
		}
		if msg.Type == tea.KeyCtrlL {
			m.clearScreen()
			return m, nil
		}
		if msg.Type == tea.KeyCtrlO {
			m.verbose = !m.verbose
			m.refreshViewport()
			return m, nil
		}
		if msg.Type == tea.KeyCtrlY {
			m.handleCopyCommand(nil)
			m.refreshViewport()
			return m, nil
		}
		if msg.Type == tea.KeyCtrlJ && !m.busy && !m.modalActive() {
			return m, m.insertNewline()
		}
		if msg.Type == tea.KeyEnd || msg.Type == tea.KeyCtrlG {
			m.viewport.GotoBottom()
			m.followOutput = true
			m.refreshViewport()
			return m, nil
		}
		if !m.busy {
			if handled, model, cmd := m.handleMentionMenuKey(msg); handled {
				return model, cmd
			}
			if handled, model, cmd := m.handleSlashMenuKey(msg); handled {
				return model, cmd
			}
		}
		if m.handleNavigationKey(msg) {
			m.refreshViewport()
			return m, nil
		}
		if m.handleHistoryKey(msg.Type) {
			return m, nil
		}
		if m.handleViewportScroll(msg) {
			m.refreshViewport()
			return m, nil
		}
		switch msg.Type {
		case tea.KeyEnter:
			if m.busy {
				return m, nil
			}
			if m.shouldInsertNewline(msg) {
				return m, m.insertNewline()
			}
			return m.submitInput()
		case tea.KeyCtrlU:
			m.input.Reset()
			m.historyIndex = -1
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	prevLines := m.slashMenuLineCount() + m.mentionMenuLineCount()
	m.syncSlashMenu()
	m.syncMentionMenu()
	newLines := m.slashMenuLineCount() + m.mentionMenuLineCount()
	if newLines != prevLines {
		m.adjustLayout()
	}
	return m, cmd
}

func (m Model) View() string {
	if m.trustPending {
		return m.renderTrustScreen()
	}
	if m.modalActive() {
		return m.renderModalScreen()
	}
	if !m.ready {
		return m.styles.subtitle.Render("\n  初始化 OpenTMD…\n")
	}

	parts := []string{m.viewport.View(), m.buildInputArea()}
	if footer := m.buildFooter(); footer != "" {
		parts = append(parts, footer)
	}
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (m *Model) clearScreen() {
	m.messages = nil
	m.followOutput = true
	m.viewport.GotoTop()
	m.refreshViewport()
}

func (m *Model) adjustLayout() {
	if m.width == 0 || m.height == 0 {
		return
	}
	slashLines := m.slashMenuLineCount()
	mentionLines := m.mentionMenuLineCount()
	menuLines := slashLines
	if mentionLines > menuLines {
		menuLines = mentionLines
	}
	inputLines := 4 + menuLines
	vpHeight := m.height - inputLines
	if vpHeight < 8 {
		vpHeight = 8
	}
	m.viewport.Width = m.width
	m.viewport.Height = vpHeight
	m.input.SetWidth(m.width - 8)
}

func (m *Model) refreshViewport() {
	offset := m.viewport.YOffset
	atBottom := m.viewport.AtBottom()
	width := m.contentWidth()
	m.rebuildMessageAnchors(width)
	m.viewport.SetContent(m.renderMessages(width))
	if m.followOutput || atBottom {
		m.viewport.GotoBottom()
	} else {
		m.viewport.YOffset = offset
		if m.viewport.PastBottom() {
			m.viewport.GotoBottom()
		}
	}
}

func (m *Model) handleViewportScroll(msg tea.Msg) bool {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if (m.slashMenuVisible() || m.mentionMenuVisible()) && (msg.Type == tea.KeyUp || msg.Type == tea.KeyDown) {
			return false
		}
		if msg.Type == tea.KeyUp || msg.Type == tea.KeyDown {
			return false
		}
		if !isScrollKey(msg) {
			return false
		}
	case tea.MouseMsg:
		if msg.Action != tea.MouseActionPress {
			return false
		}
		if msg.Button != tea.MouseButtonWheelUp && msg.Button != tea.MouseButtonWheelDown {
			return false
		}
	default:
		return false
	}

	m.viewport, _ = m.viewport.Update(msg)
	m.followOutput = m.viewport.AtBottom()
	return true
}

func isScrollKey(msg tea.KeyMsg) bool {
	switch msg.Type {
	case tea.KeyUp, tea.KeyDown, tea.KeyPgUp, tea.KeyPgDown, tea.KeyHome, tea.KeyEnd:
		return true
	case tea.KeyCtrlU, tea.KeyCtrlD:
		return true
	}
	return false
}

func isQuitKey(msg tea.KeyMsg) bool {
	return msg.Type == tea.KeyCtrlC
}

func isClearInputKey(msg tea.KeyMsg) bool {
	return msg.Type == tea.KeyEsc
}

func (m *Model) handleSlash(input string) (tea.Model, tea.Cmd) {
	parts := strings.Fields(input)
	cmd := parts[0]

	switch cmd {
	case "/help":
		m.appendSystem(strings.Join([]string{
			"可用命令:",
			"  /help              显示帮助",
			"  /clear             清屏（保留对话上下文）",
			"  /session           开始新会话（清除对话）",
			"  /resume            恢复历史 session",
			"  /session list      列出历史 session",
			"  /session <id>      切换到指定 session",
			"  /provider          Provider 管理向导（API Key / Base URL）",
			"  /login             配置 Provider 与 API Key",
			"  /cost              显示 token 用量与估算费用",
			"  /plan              切换到只读探索模式",
			"  /build             切换到可执行构建模式",
			"  /undo              撤销上一轮文件修改",
			"  /auto-run          切换自动执行（on/off/status）",
			"  /bg                后台 session 管理",
			"  /cd <dir>          切换工作目录",
			"  /status            当前 provider / 目录 / session / LSP",
			"  /config            配置文件路径与摘要",
			"  /reload            重新加载 config.toml",
			"  /lsp               显示 LSP 语言服务器状态",
			"  /lsp reload        重启 LSP 客户端",
			"  /skills            列出 Skills",
			"  /mcp               显示 MCP 服务器状态",
			"  /mcp reload        重连 MCP",
			"  /copy              复制最近 AI 回复（无回复时复制最后一条输出）",
			"  /copy all          复制完整会话",
			"  /export [file]     导出会话到文件",
			"  /keys              显示快捷键",
			"  /resume            恢复历史 session",
			"  /remember <text>   记忆一条事实（全局）",
			"  /remember project <text>  记忆一条项目相关事实",
			"  /forget <keyword>  删除匹配的记忆",
			"  /memory            查看所有记忆",
			"  /compact [focus]   压缩对话历史（LLM 摘要）",
			"  /quit              退出",
			"  /exit              同 /quit",
			"",
			"输入 @ 可补全项目文件路径",
			"",
			"快捷键:",
			"  Enter          发送",
			"  Ctrl+J / Alt+Enter  换行",
			"  Tab            斜杠命令补全",
			"  Esc            取消输出 / 清空输入",
			"  Ctrl+O         切换工具详情显示",
			"  Ctrl+L         清屏（保留对话）",
			"  Ctrl+G / End   滚到最新",
			"  Ctrl+C         取消；连按两次退出",
			"",
			"提示: SSH 无图形环境时可用鼠标选中终端历史输出；/copy 会写入 ~/.opentmd/last-copy.txt",
			"",
			"权限:  y/Y 允许  a/A 会话允许  n/N 拒绝",
		}, "\n"))
		m.refreshViewport()
		return m, nil

	case "/clear":
		m.clearScreen()
		m.refreshViewport()
		return m, nil

	case "/cost":
		u := m.agent.Usage()
		cfg := m.agent.Config()
		cost := pricing.EstimateCost(cfg.Default.Provider, cfg.Default.Model, u.PromptTokens, u.CompletionTokens)
		m.appendSystem(strings.Join([]string{
			fmt.Sprintf("Token 用量（当前 session）:"),
			fmt.Sprintf("  输入:   %d", u.PromptTokens),
			fmt.Sprintf("  输出:   %d", u.CompletionTokens),
			fmt.Sprintf("  合计:   %d", u.TotalTokens),
			fmt.Sprintf("  估算:   %s", pricing.FormatCost(cost)),
		}, "\n"))
		m.refreshViewport()
		return m, nil

	case "/undo":
		paths, err := m.agent.UndoLastTurn()
		if err != nil {
			m.appendSystem("✗ 撤销失败: " + err.Error())
		} else {
			var rel []string
			for _, p := range paths {
				rel = append(rel, filehistory.RelDisplay(m.agent.WorkDir(), p))
			}
			m.appendSystem("✓ 已恢复 " + fmt.Sprintf("%d", len(rel)) + " 个文件:\n  • " + strings.Join(rel, "\n  • "))
		}
		m.refreshViewport()
		return m, nil

	case "/plan":
		m.agent.SetPlanMode(true)
		m.appendSystem("✓ 已切换到 /plan：只读探索模式（禁用写文件、bash、MCP）")
		m.refreshViewport()
		return m, nil

	case "/build":
		m.agent.SetPlanMode(false)
		m.appendSystem("✓ 已切换到 /build：可执行构建模式")
		m.refreshViewport()
		return m, nil

	case "/login":
		m.openLoginWizard()
		m.refreshViewport()
		return m, nil

	case "/auto-run":
		m.handleAutoRun(parts[1:])
		m.refreshViewport()
		return m, nil

	case "/bg":
		m.handleBgCommand(parts[1:])
		m.refreshViewport()
		return m, nil

	case "/keys":
		m.appendSystem(keybindingsHelp())
		m.refreshViewport()
		return m, nil

	case "/resume":
		m.openSessionPicker()
		m.refreshViewport()
		return m, nil

	case "/quit":
		return m, tea.Quit

	case "/session":
		if len(parts) == 1 {
			if err := m.agent.NewSession(); err != nil {
				m.appendSystem("✗ 新建 session 失败: " + err.Error())
			} else {
				m.messages = nil
				m.agent.ResetUsage()
				m.bgSessions = nil
				m.appendSystem("✓ 已开始新会话")
			}
			m.refreshViewport()
			return m, nil
		}
		return m.handleSessionCommand(parts[1:])

	case "/model":
		m.openModelPicker()
		m.refreshViewport()
		return m, nil

	case "/provider":
		m.openProviderWizard()
		m.refreshViewport()
		return m, nil

	case "/cd":
		if len(parts) < 2 {
			m.appendSystem("用法: /cd <目录>")
			m.refreshViewport()
			return m, nil
		}
		dir := parts[1]
		if err := m.agent.SetWorkDir(dir); err != nil {
			m.appendSystem("✗ 切换目录失败: " + err.Error())
		} else {
			m.appendSystem("✓ 工作目录: " + m.agent.WorkDir())
			m.requireTrustForDir(m.agent.WorkDir())
			m.invalidateFileIndex()
		}
		m.refreshViewport()
		return m, nil

	case "/reload":
		cfg, err := config.Load()
		if err != nil {
			m.appendSystem("✗ 加载配置失败: " + err.Error())
		} else if err := m.agent.Reload(cfg); err != nil {
			m.appendSystem("✗ 重载失败: " + err.Error())
		} else {
			m.theme = loadTheme(cfg.TUI.Theme)
			m.styles = m.theme.styles()
			m.viewport.Style = m.styles.viewport
			applyInputStyles(&m.input, m.theme)
			m.appendSystem(fmt.Sprintf("✓ 已重载: %s / %s", cfg.Default.Provider, cfg.Default.Model))
		}
		m.refreshViewport()
		return m, nil

	case "/status":
		sess := m.agent.Session().Current()
		cfg := m.agent.Config()
		msgCount := 0
		sessID := "(none)"
		title := ""
		if sess != nil {
			sessID = sess.ID
			title = sess.Title
			msgCount = len(sess.Messages)
		}
		lines := []string{
			fmt.Sprintf("Provider: %s / %s", cfg.Default.Provider, cfg.Default.Model),
			fmt.Sprintf("Workdir:  %s", m.agent.WorkDir()),
			fmt.Sprintf("Session:  %s (%d 条消息)", sessID, msgCount),
		}
		if m.agent.PlanMode() {
			lines = append(lines, "Mode:     plan (read-only)")
		} else {
			lines = append(lines, "Mode:     build (execution enabled)")
		}
		if title != "" {
			lines = append(lines, fmt.Sprintf("Title:    %s", title))
		}
		lines = append(lines, m.agent.LSPStatus())
		if m.autoRun {
			lines = append(lines, "Auto-run: 开启（常规工具自动执行）")
		}
		m.appendSystem(strings.Join(lines, "\n"))
		m.refreshViewport()
		return m, nil

	case "/config":
		path, err := config.Path()
		if err != nil {
			m.appendSystem("✗ 配置路径: " + err.Error())
		} else {
			cfg := m.agent.Config()
			m.appendSystem(strings.Join([]string{
				fmt.Sprintf("Config: %s", path),
				fmt.Sprintf("  provider: %s", cfg.Default.Provider),
				fmt.Sprintf("  model:    %s", cfg.Default.Model),
				fmt.Sprintf("  theme:    %s", cfg.TUI.Theme),
				fmt.Sprintf("  stream:   %v", cfg.TUI.Stream),
				fmt.Sprintf("  persist:  %v", cfg.Session.Persist),
				fmt.Sprintf("  lsp:      enabled=%v auto_detect=%v settle_ms=%d",
					cfg.LSP.Enabled, cfg.LSP.AutoDetect, cfg.LSP.DiagnosticsSettleDelayMs),
			}, "\n"))
		}
		m.refreshViewport()
		return m, nil

	case "/lsp":
		if len(parts) >= 2 && parts[1] == "reload" {
			m.agent.ReloadLSP()
			m.appendSystem("✓ LSP 已重启")
		} else {
			m.appendSystem(m.agent.LSPStatus())
		}
		m.refreshViewport()
		return m, nil

	case "/skills":
		names := m.agent.ListSkills()
		if len(names) == 0 {
			m.appendSystem("未找到 Skills")
		} else {
			m.appendSystem("Skills:\n  • " + strings.Join(names, "\n  • "))
		}
		m.refreshViewport()
		return m, nil

	case "/mcp":
		if len(parts) >= 2 && parts[1] == "reload" {
			if err := m.agent.ReloadMCP(context.Background()); err != nil {
				m.appendSystem("✗ MCP 重载失败: " + err.Error())
			} else {
				m.appendSystem("✓ MCP 已重载")
			}
		} else {
			m.appendSystem(m.agent.MCPStatus(context.Background()).Text)
		}
		m.refreshViewport()
		return m, nil

	case "/copy":
		m.handleCopyCommand(parts[1:])
		m.refreshViewport()
		return m, nil

	case "/export":
		m.handleExportCommand(parts[1:])
		m.refreshViewport()
		return m, nil

	case "/exit":
		return m, tea.Quit

	case "/remember":
		m.handleRemember(parts[1:])
		m.refreshViewport()
		return m, nil

	case "/forget":
		m.handleForget(parts[1:])
		m.refreshViewport()
		return m, nil

	case "/memory":
		m.handleMemory()
		m.refreshViewport()
		return m, nil

	case "/compact":
		focus := strings.TrimSpace(strings.Join(parts[1:], " "))
		if m.chatCancel != nil {
			m.chatCancel()
		}
		ctx, cancel := context.WithCancel(context.Background())
		m.chatCancel = cancel
		m.busy = true
		m.appendSystem("正在压缩对话历史…")
		m.refreshViewport()
		return m, m.runCompact(ctx, focus)

	default:
		m.appendSystem("未知命令: " + cmd + "，输入 /help 查看帮助")
		m.refreshViewport()
		return m, nil
	}
}

func (m *Model) handleSessionCommand(args []string) (tea.Model, tea.Cmd) {
	if len(args) == 0 {
		sess := m.agent.Session().Current()
		m.appendSystem(fmt.Sprintf(
			"Session: %s\n标题: %s\n消息: %d 条",
			sess.ID, sess.Title, len(sess.Messages),
		))
		m.refreshViewport()
		return m, nil
	}

	switch args[0] {
	case "list":
		sessions, err := m.agent.ListSessions()
		if err != nil {
			m.appendSystem("✗ 列出 session 失败: " + err.Error())
			m.refreshViewport()
			return m, nil
		}
		var lines []string
		lines = append(lines, "历史 Sessions:")
		current := m.agent.Session().Current().ID
		for i, s := range sessions {
			if i >= 10 {
				lines = append(lines, fmt.Sprintf("  … 共 %d 个", len(sessions)))
				break
			}
			marker := " "
			if s.ID == current {
				marker = "●"
			}
			idShort := s.ID
			if len(idShort) > 8 {
				idShort = idShort[:8]
			}
			lines = append(lines, fmt.Sprintf("  %s [%s] %s (%d)", marker, idShort, s.Title, len(s.Messages)))
		}
		m.appendSystem(strings.Join(lines, "\n"))
		m.refreshViewport()
		return m, nil

	case "new":
		if err := m.agent.NewSession(); err != nil {
			m.appendSystem("✗ 新建失败: " + err.Error())
		} else {
			m.loadSessionMessages()
			m.agent.ResetUsage()
			m.appendSystem("✓ 已创建新 session")
		}
		m.refreshViewport()
		return m, nil

	default:
		prefix := args[0]
		sessions, err := m.agent.ListSessions()
		if err != nil {
			m.appendSystem("✗ 查找失败: " + err.Error())
			m.refreshViewport()
			return m, nil
		}
		var matches []string
		for _, s := range sessions {
			if strings.HasPrefix(s.ID, prefix) {
				matches = append(matches, s.ID)
			}
		}
		if len(matches) == 0 {
			m.appendSystem("未找到 session: " + prefix)
			m.refreshViewport()
			return m, nil
		}
		if len(matches) > 1 {
			m.appendSystem("匹配多个 session，请输入更长前缀")
			m.refreshViewport()
			return m, nil
		}
		if err := m.agent.LoadSession(matches[0]); err != nil {
			m.appendSystem("✗ 加载失败: " + err.Error())
		} else {
			m.loadSessionMessages()
			m.agent.ResetUsage()
			m.appendSystem("✓ 已切换: " + matches[0][:8])
		}
		m.refreshViewport()
		return m, nil
	}
}

func (m *Model) handleRemember(args []string) {
	if len(args) == 0 {
		m.appendSystem("用法: /remember <内容> 或 /remember project <内容>")
		return
	}
	scope := "global"
	content := strings.Join(args, " ")
	if len(args) >= 2 && args[0] == "project" {
		scope = "project"
		content = strings.Join(args[1:], " ")
	}
	if err := m.agent.Remember(scope, content); err != nil {
		m.appendSystem("✗ 记忆失败: " + err.Error())
		return
	}
	m.appendSystem("✓ 已记忆（" + scope + "）: " + content)
}

func (m *Model) handleForget(args []string) {
	if len(args) == 0 {
		m.appendSystem("用法: /forget <关键字> 或 /forget project <关键字>")
		return
	}
	scope := "global"
	keyword := strings.Join(args, " ")
	if len(args) >= 2 && args[0] == "project" {
		scope = "project"
		keyword = strings.Join(args[1:], " ")
	}
	removed, err := m.agent.Forget(scope, keyword)
	if err != nil {
		m.appendSystem("✗ 删除记忆失败: " + err.Error())
		return
	}
	if len(removed) == 0 {
		m.appendSystem("未找到包含「" + keyword + "」的记忆")
		return
	}
	m.appendSystem("✓ 已删除 " + fmt.Sprintf("%d", len(removed)) + " 条记忆:\n  • " + strings.Join(removed, "\n  • "))
}

func (m *Model) handleMemory() {
	entries := m.agent.MemoryList("all")
	if len(entries) == 0 {
		m.appendSystem("暂无记忆。使用 /remember <内容> 添加。")
		return
	}
	m.appendSystem("记忆:\n  • " + strings.Join(entries, "\n  • "))
}

func (m Model) streamChat(ctx context.Context, input string) tea.Cmd {
	return func() tea.Msg {
		// Run the agent outside the Cmd worker so bubbletea's command loop
		// stays responsive while tool status / stream chunks are delivered.
		go func() {
			content, err := m.agent.ChatWithStatus(ctx, input, func(chunk string) error {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				if m.program != nil {
					m.program.Send(chunkMsg{content: chunk})
				}
				return nil
			}, func(status string) {
				if ctx.Err() != nil {
					return
				}
				if m.program != nil {
					m.program.Send(statusMsg{text: status})
				}
			})
			if m.program != nil {
				m.program.Send(doneMsg{err: err, content: content})
			}
		}()
		return nil
	}
}

func (m Model) runCompact(ctx context.Context, focus string) tea.Cmd {
	return func() tea.Msg {
		go func() {
			result, err := m.agent.CompactWithStatus(ctx, focus, func(status string) {
				if ctx.Err() != nil {
					return
				}
				if m.program != nil {
					m.program.Send(statusMsg{text: status})
				}
			})
			if m.program != nil {
				m.program.Send(compactDoneMsg{text: result, err: err})
			}
		}()
		return nil
	}
}

func Run(ag *agent.Agent, trustOpts trust.Options) error {
	defer ag.Shutdown()

	approvalCh := make(chan permission.Decision, 1)
	m := New(ag, trustOpts)
	m.approvalCh = approvalCh

	opts := []tea.ProgramOption{tea.WithMouseAllMotion()}
	if useAltScreen() {
		opts = append(opts, tea.WithAltScreen())
	}
	p := tea.NewProgram(&m, opts...)
	m.program = p

	ag.SetApprovalHandler(m.resolveApproval)
	ag.SetLSPEventHandler(func(ev lsp.ConnectEvent) {
		if m.program != nil {
			m.program.Send(lspConnectMsg{ev: ev})
		}
	})

	_, err := p.Run()
	return err
}
