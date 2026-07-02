package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type messageAnchor struct {
	index     int
	lineStart int
	kind      string
}

func (m *Model) rebuildMessageAnchors(contentWidth int) {
	m.messageAnchors = nil
	if len(m.messages) == 0 {
		return
	}
	line := 0
	lastAI := m.lastAssistantIdx()
	for i, msg := range m.messages {
		m.messageAnchors = append(m.messageAnchors, messageAnchor{
			index:     i,
			lineStart: line,
			kind:      msg.kind,
		})
		rendered := m.renderOneMessage(msg, contentWidth, i, lastAI)
		line += strings.Count(rendered, "\n") + 1
	}
}

func (m *Model) scrollLines(delta int) {
	m.viewport.YOffset += delta
	if m.viewport.YOffset < 0 {
		m.viewport.YOffset = 0
	}
	m.followOutput = m.viewport.AtBottom()
}

func (m *Model) jumpMessage(userOnly bool, dir int) {
	if len(m.messageAnchors) == 0 {
		return
	}
	current := m.viewport.YOffset
	idx := 0
	for i := len(m.messageAnchors) - 1; i >= 0; i-- {
		if m.messageAnchors[i].lineStart <= current {
			idx = i
			break
		}
	}
	for {
		idx += dir
		if idx < 0 || idx >= len(m.messageAnchors) {
			return
		}
		if !userOnly || m.messageAnchors[idx].kind == "user" {
			m.viewport.YOffset = m.messageAnchors[idx].lineStart
			m.followOutput = false
			return
		}
	}
}

func (m *Model) handleNavigationKey(msg tea.KeyMsg) bool {
	switch msg.Type {
	case tea.KeyShiftUp:
		m.scrollLines(-1)
		return true
	case tea.KeyShiftDown:
		m.scrollLines(1)
		return true
	case tea.KeyUp:
		if msg.Alt {
			m.jumpMessage(false, -1)
			return true
		}
	case tea.KeyDown:
		if msg.Alt {
			m.jumpMessage(false, 1)
			return true
		}
	case tea.KeyCtrlUp:
		m.jumpMessage(true, -1)
		return true
	case tea.KeyCtrlDown:
		m.jumpMessage(true, 1)
		return true
	}
	return false
}
