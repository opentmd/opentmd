package tui

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/opentmd/opentmd-cli/internal/config"
)

type modalKind int

const (
	modalNone modalKind = iota
	modalModelPicker
	modalSessionPicker
	modalProviderMenu
	modalProviderPick
	modalProviderAdd
	modalTextInput
	modalLoginProvider
	modalLoginAPIKey
	modalLoginModel
)

type pickerItem struct {
	Label  string
	Value  string
	Detail string
}

func (m *Model) openModelPicker() {
	cfg := m.agent.Config()
	current := cfg.Default.Provider
	var items []pickerItem
	for _, p := range config.ProviderPresets {
		pc := cfg.Providers[p.Name]
		if p.NeedAPIKey && pc.APIKey == "" {
			continue
		}
		model := cfg.Default.Model
		if p.Name != current {
			model = p.DefaultModel
		}
		items = append(items, pickerItem{
			Label:  fmt.Sprintf("%s / %s", p.Label, model),
			Value:  p.Name,
			Detail: p.Name,
		})
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Value == current {
			return true
		}
		if items[j].Value == current {
			return false
		}
		return items[i].Label < items[j].Label
	})
	m.modalKind = modalModelPicker
	m.modalItems = items
	m.modalSelected = 0
}

func (m *Model) openSessionPicker() {
	sessions, err := m.agent.ListSessions()
	if err != nil {
		m.appendSystem("✗ 列出 session 失败: " + err.Error())
		return
	}
	current := ""
	if s := m.agent.Session().Current(); s != nil {
		current = s.ID
	}
	var items []pickerItem
	for _, s := range sessions {
		idShort := s.ID
		if len(idShort) > 8 {
			idShort = idShort[:8]
		}
		label := fmt.Sprintf("[%s] %s", idShort, s.Title)
		if s.ID == current {
			label = "● " + label
		}
		items = append(items, pickerItem{
			Label:  label,
			Value:  s.ID,
			Detail: fmt.Sprintf("%d messages", len(s.Messages)),
		})
	}
	m.modalKind = modalSessionPicker
	m.modalItems = items
	m.modalSelected = 0
}

func (m *Model) closeModal() {
	m.modalKind = modalNone
	m.modalItems = nil
	m.modalSelected = 0
	m.modalTextBuf = ""
	m.modalTextAction = ""
	m.modalTextTarget = ""
	m.modalTextHint = ""
	m.modalTextMask = false
	m.loginPreset = ""
	m.loginAPIKey = ""
}

func (m *Model) modalActive() bool {
	return m.modalKind != modalNone
}

func (m *Model) handleModalKey(msg tea.KeyMsg) bool {
	if !m.modalActive() {
		return false
	}
	if m.modalKind == modalTextInput {
		return m.handleTextInputKey(msg)
	}
	if m.modalKind == modalLoginAPIKey || m.modalKind == modalLoginModel {
		return m.handleLoginTextInputKey(msg)
	}
	switch msg.Type {
	case tea.KeyEsc:
		m.closeModal()
		return true
	case tea.KeyUp, tea.KeyCtrlP:
		m.modalSelected = cycleSlashSelection(m.modalSelected, len(m.modalItems), true)
		return true
	case tea.KeyDown, tea.KeyCtrlN:
		m.modalSelected = cycleSlashSelection(m.modalSelected, len(m.modalItems), false)
		return true
	case tea.KeyEnter:
		if len(m.modalItems) == 0 {
			m.closeModal()
			return true
		}
		item := m.modalItems[m.modalSelected]
		switch m.modalKind {
		case modalModelPicker:
			m.applyModelChoice(item.Value)
			m.closeModal()
		case modalSessionPicker:
			m.applySessionChoice(item.Value)
			m.closeModal()
		case modalProviderMenu, modalProviderPick, modalProviderAdd:
			m.handleProviderModalEnter(item)
		case modalLoginProvider:
			m.handleLoginModalEnter(item)
		}
		return true
	}
	return false
}

func (m *Model) applyModelChoice(provider string) {
	cfg := m.agent.Config()
	if p, ok := config.PresetByName(provider); ok {
		cfg.ApplyPreset(p)
	} else {
		cfg.Default.Provider = provider
	}
	if err := config.Save(cfg); err != nil {
		m.appendSystem("✗ 保存配置失败: " + err.Error())
		return
	}
	if err := m.agent.Reload(cfg); err != nil {
		m.appendSystem("✗ 切换 provider 失败: " + err.Error())
		return
	}
	m.appendSystem(fmt.Sprintf("✓ 已切换: %s / %s", cfg.Default.Provider, cfg.Default.Model))
	m.refreshViewport()
}

func (m *Model) applySessionChoice(id string) {
	if err := m.agent.LoadSession(id); err != nil {
		m.appendSystem("✗ 加载 session 失败: " + err.Error())
		return
	}
	m.loadSessionMessages()
	m.agent.ResetUsage()
	short := id
	if len(short) > 8 {
		short = short[:8]
	}
	m.appendSystem("✓ 已切换 session: " + short)
	m.refreshViewport()
}

func (m *Model) renderModal() string {
	if !m.modalActive() {
		return ""
	}
	if m.modalKind == modalTextInput {
		return m.renderTextInputModal()
	}
	if m.modalKind == modalLoginAPIKey || m.modalKind == modalLoginModel {
		return m.renderLoginModal()
	}
	title := "选择 Provider / Model"
	switch m.modalKind {
	case modalSessionPicker:
		title = "选择 Session"
	case modalProviderMenu:
		title = "Provider 管理"
	case modalProviderPick:
		title = "选择 Provider"
	case modalProviderAdd:
		title = "添加 Provider（预设）"
	case modalLoginProvider:
		title = "登录 — 选择 Provider"
	}
	if len(m.modalItems) == 0 {
		return m.renderModalBox(title, m.styles.hint.Render("  (无可用项)"))
	}

	selected := m.modalSelected
	if selected >= len(m.modalItems) {
		selected = 0
	}
	start, end, _ := slashMenuWindow(selected, len(m.modalItems), slashMenuMaxVisible, 0)

	active := lipgloss.NewStyle().Foreground(m.theme.Assistant).Bold(true)
	muted := lipgloss.NewStyle().Foreground(m.theme.Muted)
	var lines []string
	for i := start; i < end; i++ {
		item := m.modalItems[i]
		prefix := "  "
		nameStyle := muted
		if i == selected {
			prefix = sym.Arrow
			nameStyle = active
		}
		line := prefix + nameStyle.Render(item.Label)
		if item.Detail != "" {
			line += "  " + muted.Render(item.Detail)
		}
		lines = append(lines, line)
	}
	body := strings.Join(lines, "\n")
	hint := m.styles.inputBar.Render(sym.Up + sym.Down + " 循环 · Enter 确认 · Esc 取消")
	return m.renderModalBox(title, body+"\n"+hint)
}

func (m *Model) renderModalBox(title, body string) string {
	width := min(m.width-6, 72)
	if width < 40 {
		width = 40
	}
	content := lipgloss.NewStyle().Bold(true).Render(title) + "\n\n" + body
	return lipgloss.NewStyle().
		Border(borderFor(termCaps)).
		BorderForeground(m.theme.Primary).
		Padding(1, 2).
		Width(width).
		Render(content)
}

func (m *Model) renderModalScreen() string {
	if !m.ready {
		return sym.Normalize(m.styles.subtitle.Render("\n  初始化 OpenTMD" + sym.Ellipsis + "\n"))
	}
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		m.renderModal(),
		lipgloss.WithWhitespaceChars(" "),
	)
}
