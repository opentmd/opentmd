package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/opentmd/opentmd/internal/config"
)

func (m *Model) openLoginWizard() {
	var items []pickerItem
	for _, p := range config.ProviderPresets {
		items = append(items, pickerItem{
			Label:  fmt.Sprintf("%s — %s", p.Label, p.DefaultModel),
			Value:  p.Name,
			Detail: p.BaseURL,
		})
	}
	m.modalKind = modalLoginProvider
	m.modalItems = items
	m.modalSelected = 0
	m.modalTextBuf = ""
	m.loginPreset = ""
	m.loginAPIKey = ""
}

func (m *Model) beginLoginAPIKey(name string) {
	preset, ok := config.PresetByName(name)
	if !ok {
		m.appendSystem("✗ 未知 provider")
		m.closeModal()
		return
	}
	m.loginPreset = name
	if !preset.NeedAPIKey {
		m.finishLogin(name, "", preset.DefaultModel)
		return
	}
	m.modalKind = modalLoginAPIKey
	m.modalItems = nil
	m.modalTextBuf = ""
	m.modalTextMask = true
	m.modalTextHint = preset.Label + " — 请输入 API Key"
}

func (m *Model) beginLoginModel() {
	preset, _ := config.PresetByName(m.loginPreset)
	m.modalKind = modalLoginModel
	m.modalTextBuf = preset.DefaultModel
	m.modalTextMask = false
	m.modalTextHint = "默认模型（Enter 使用显示值）"
}

func (m *Model) finishLogin(provider, apiKey, model string) {
	cfg := m.agent.Config()
	preset, ok := config.PresetByName(provider)
	if !ok {
		m.appendSystem("✗ 未知 provider")
		return
	}
	cfg.ApplyPreset(preset)
	if apiKey != "" {
		_ = cfg.SetAPIKey(provider, apiKey)
	} else if !preset.NeedAPIKey {
		_ = cfg.SetAPIKey(provider, "ollama")
	}
	if model != "" {
		cfg.Default.Model = model
	}
	if err := config.Save(cfg); err != nil {
		m.appendSystem("✗ 保存配置失败: " + err.Error())
		return
	}
	_ = cfg.MarkSetupCompleted()
	if err := m.agent.Reload(cfg); err != nil {
		m.appendSystem("✗ 重载失败: " + err.Error())
		return
	}
	m.closeModal()
	m.appendSystem(fmt.Sprintf("✓ 登录完成: %s / %s", cfg.Default.Provider, cfg.Default.Model))
	m.refreshViewport()
}

func (m *Model) handleLoginModalEnter(item pickerItem) {
	if m.modalKind == modalLoginProvider {
		m.beginLoginAPIKey(item.Value)
	}
}

func (m *Model) handleLoginTextInputKey(msg tea.KeyMsg) bool {
	switch msg.Type {
	case tea.KeyEsc:
		m.closeModal()
		return true
	case tea.KeyEnter:
		switch m.modalKind {
		case modalLoginAPIKey:
			key := strings.TrimSpace(m.modalTextBuf)
			if key == "" {
				return true
			}
			m.loginAPIKey = key
			m.beginLoginModel()
		case modalLoginModel:
			model := strings.TrimSpace(m.modalTextBuf)
			m.finishLogin(m.loginPreset, m.loginAPIKey, model)
		}
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

func (m *Model) renderLoginModal() string {
	switch m.modalKind {
	case modalLoginAPIKey:
		display := strings.Repeat("•", len(m.modalTextBuf))
		body := m.modalTextHint + "\n\n" + display + "▌"
		hint := m.styles.inputBar.Render("Enter 下一步 · Esc 取消")
		return m.renderModalBox("登录 — API Key", body+"\n"+hint)
	case modalLoginModel:
		body := m.modalTextHint + "\n\n" + m.modalTextBuf + "▌"
		hint := m.styles.inputBar.Render("Enter 完成 · Esc 取消")
		return m.renderModalBox("登录 — 模型", body+"\n"+hint)
	}
	return ""
}
