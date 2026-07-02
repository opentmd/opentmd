package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/opentmd/opentmd/internal/config"
)

func (m *Model) openProviderWizard() {
	m.modalKind = modalProviderMenu
	m.modalItems = []pickerItem{
		{Label: "切换默认 Provider / Model", Value: "switch"},
		{Label: "配置 API Key", Value: "apikey"},
		{Label: "设置 Base URL", Value: "baseurl"},
		{Label: "从预设添加 Provider", Value: "add"},
		{Label: "查看已配置 Provider", Value: "list"},
	}
	m.modalSelected = 0
	m.modalTextBuf = ""
}

func (m *Model) openProviderPick(action string) {
	cfg := m.agent.Config()
	var items []pickerItem
	names := make(map[string]bool)
	for name := range cfg.Providers {
		names[name] = true
	}
	names[cfg.Default.Provider] = true
	for name := range names {
		label := name
		if p, ok := config.PresetByName(name); ok {
			label = p.Label + " (" + name + ")"
		}
		if name == cfg.Default.Provider {
			label = "● " + label
		}
		items = append(items, pickerItem{Label: label, Value: name})
	}
	if len(items) == 0 {
		for _, p := range config.ProviderPresets {
			items = append(items, pickerItem{Label: p.Label, Value: p.Name})
		}
	}
	m.modalKind = modalProviderPick
	m.modalTextAction = action
	m.modalItems = items
	m.modalSelected = 0
}

func (m *Model) openProviderPresetAdd() {
	var items []pickerItem
	for _, p := range config.ProviderPresets {
		items = append(items, pickerItem{
			Label:  fmt.Sprintf("%s — %s", p.Label, p.DefaultModel),
			Value:  p.Name,
			Detail: p.BaseURL,
		})
	}
	m.modalKind = modalProviderAdd
	m.modalItems = items
	m.modalSelected = 0
}

func (m *Model) showProviderList() {
	cfg := m.agent.Config()
	var lines []string
	lines = append(lines, "已配置 Providers:")
	for name, pc := range cfg.Providers {
		marker := " "
		if name == cfg.Default.Provider {
			marker = "●"
		}
		keyStatus := "未设置"
		if pc.APIKey != "" {
			keyStatus = "已设置"
		}
		lines = append(lines, fmt.Sprintf("  %s %s  base=%s  key=%s", marker, name, pc.BaseURL, keyStatus))
	}
	lines = append(lines, fmt.Sprintf("默认: %s / %s", cfg.Default.Provider, cfg.Default.Model))
	m.appendSystem(strings.Join(lines, "\n"))
}

func (m *Model) handleProviderMenuChoice(value string) {
	switch value {
	case "switch":
		m.openModelPicker()
	case "apikey":
		m.openProviderPick("apikey")
	case "baseurl":
		m.openProviderPick("baseurl")
	case "add":
		m.openProviderPresetAdd()
	case "list":
		m.closeModal()
		m.showProviderList()
	}
}

func (m *Model) beginProviderTextInput(action, provider string) {
	m.modalKind = modalTextInput
	m.modalTextAction = action
	m.modalTextTarget = provider
	m.modalTextBuf = ""
	m.modalTextMask = action == "apikey"
	title := "Base URL"
	if action == "apikey" {
		title = "API Key"
	}
	if p, ok := config.PresetByName(provider); ok {
		m.modalTextHint = fmt.Sprintf("%s (%s)", p.Label, provider)
		if action == "baseurl" && p.BaseURL != "" {
			m.modalTextHint += "  默认: " + p.BaseURL
		}
	} else {
		m.modalTextHint = provider
	}
	_ = title
}

func (m *Model) commitProviderTextInput() {
	cfg := m.agent.Config()
	if cfg.Providers == nil {
		cfg.Providers = map[string]config.ProviderConfig{}
	}
	pc := cfg.Providers[m.modalTextTarget]
	val := strings.TrimSpace(m.modalTextBuf)
	switch m.modalTextAction {
	case "apikey":
		if val == "" {
			m.appendSystem("✗ API Key 不能为空")
			return
		}
		pc.APIKey = val
	case "baseurl":
		if val == "" {
			m.appendSystem("✗ Base URL 不能为空")
			return
		}
		pc.BaseURL = val
	}
	cfg.Providers[m.modalTextTarget] = pc
	if err := config.Save(cfg); err != nil {
		m.appendSystem("✗ 保存失败: " + err.Error())
		return
	}
	if err := m.agent.Reload(cfg); err != nil {
		m.appendSystem("✗ 重载失败: " + err.Error())
		return
	}
	m.appendSystem(fmt.Sprintf("✓ 已更新 %s 的 %s", m.modalTextTarget, m.modalTextAction))
	m.closeModal()
	m.refreshViewport()
}

func (m *Model) applyProviderAdd(name string) {
	cfg := m.agent.Config()
	preset, ok := config.PresetByName(name)
	if !ok {
		m.appendSystem("✗ 未知 provider: " + name)
		return
	}
	cfg.ApplyPreset(preset)
	if err := config.Save(cfg); err != nil {
		m.appendSystem("✗ 保存失败: " + err.Error())
		return
	}
	if err := m.agent.Reload(cfg); err != nil {
		m.appendSystem("✗ 重载失败: " + err.Error())
		return
	}
	m.appendSystem(fmt.Sprintf("✓ 已添加 %s，当前默认: %s / %s", preset.Label, cfg.Default.Provider, cfg.Default.Model))
	if preset.NeedAPIKey {
		m.beginProviderTextInput("apikey", name)
		return
	}
	m.closeModal()
	m.refreshViewport()
}

func (m *Model) handleProviderModalEnter(item pickerItem) {
	switch m.modalKind {
	case modalProviderMenu:
		m.handleProviderMenuChoice(item.Value)
	case modalProviderPick:
		m.beginProviderTextInput(m.modalTextAction, item.Value)
	case modalProviderAdd:
		m.applyProviderAdd(item.Value)
	}
}

func (m *Model) handleTextInputKey(msg tea.KeyMsg) bool {
	switch msg.Type {
	case tea.KeyEsc:
		m.closeModal()
		return true
	case tea.KeyEnter:
		m.commitProviderTextInput()
		return true
	case tea.KeyBackspace, tea.KeyCtrlH:
		if len(m.modalTextBuf) > 0 {
			m.modalTextBuf = m.modalTextBuf[:len(m.modalTextBuf)-1]
		}
		return true
	case tea.KeySpace:
		m.modalTextBuf += " "
		return true
	case tea.KeyRunes:
		m.modalTextBuf += string(msg.Runes)
		return true
	}
	return false
}

func (m *Model) renderTextInputModal() string {
	title := "输入 API Key"
	if m.modalTextAction == "baseurl" {
		title = "输入 Base URL"
	}
	display := m.modalTextBuf
	if m.modalTextMask {
		display = strings.Repeat("•", len(m.modalTextBuf))
	}
	body := m.modalTextHint + "\n\n" + display + "▌"
	hint := m.styles.inputBar.Render("Enter 确认 · Esc 取消")
	return m.renderModalBox(title, body+"\n"+hint)
}
