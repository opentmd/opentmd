package tui

import (
	"fmt"
	"strings"
)

type bgSession struct {
	ID    string
	Title string
}

func (m *Model) handleBgCommand(args []string) {
	if len(args) == 0 {
		m.bgPushCurrent()
		return
	}
	switch args[0] {
	case "help":
		m.appendSystem(strings.Join([]string{
			"后台会话:",
			"  /bg              当前会话放到后台，新建前台会话",
			"  /bg list         列出后台会话",
			"  /bg <N>          恢复第 N 号后台会话（1-based）",
			"  /bg drop <N>     丢弃第 N 号后台会话",
		}, "\n"))
	case "list":
		m.bgList()
	default:
		if args[0] == "drop" {
			if len(args) < 2 {
				m.appendSystem("用法: /bg drop <N>")
				return
			}
			m.bgDrop(args[1])
			return
		}
		m.bgResume(args[0])
	}
}

func (m *Model) bgPushCurrent() {
	sess := m.agent.Session().Current()
	if sess == nil {
		m.appendSystem("✗ 无当前 session")
		return
	}
	if m.busy {
		m.appendSystem("✗ Agent 运行中，无法切换 session")
		return
	}
	m.bgSessions = append(m.bgSessions, bgSession{ID: sess.ID, Title: sess.Title})
	if err := m.agent.NewSession(); err != nil {
		m.appendSystem("✗ 新建 session 失败: " + err.Error())
		return
	}
	m.messages = nil
	m.agent.ResetUsage()
	idShort := sess.ID
	if len(idShort) > 8 {
		idShort = idShort[:8]
	}
	m.appendSystem(fmt.Sprintf("✓ Session [%s] 已放到后台 (#%d)，已创建新 session", idShort, len(m.bgSessions)))
	m.refreshViewport()
}

func (m *Model) bgList() {
	if len(m.bgSessions) == 0 {
		m.appendSystem("无后台 session")
		return
	}
	var lines []string
	lines = append(lines, "后台 Sessions:")
	for i, s := range m.bgSessions {
		idShort := s.ID
		if len(idShort) > 8 {
			idShort = idShort[:8]
		}
		lines = append(lines, fmt.Sprintf("  %d) [%s] %s", i+1, idShort, s.Title))
	}
	m.appendSystem(strings.Join(lines, "\n"))
}

func (m *Model) bgResume(indexStr string) {
	idx, err := parseBgIndex(indexStr, len(m.bgSessions))
	if err != nil {
		m.appendSystem("✗ " + err.Error())
		return
	}
	if m.busy {
		m.appendSystem("✗ Agent 运行中，无法切换 session")
		return
	}
	entry := m.bgSessions[idx]
	current := m.agent.Session().Current()
	if current != nil {
		m.bgSessions[idx] = bgSession{ID: current.ID, Title: current.Title}
	} else {
		m.bgSessions = append(m.bgSessions[:idx], m.bgSessions[idx+1:]...)
	}
	if err := m.agent.LoadSession(entry.ID); err != nil {
		m.appendSystem("✗ 恢复失败: " + err.Error())
		return
	}
	m.loadSessionMessages()
	m.agent.ResetUsage()
	idShort := entry.ID
	if len(idShort) > 8 {
		idShort = idShort[:8]
	}
	m.appendSystem("✓ 已恢复后台 session: [" + idShort + "]")
	m.refreshViewport()
}

func (m *Model) bgDrop(indexStr string) {
	idx, err := parseBgIndex(indexStr, len(m.bgSessions))
	if err != nil {
		m.appendSystem("✗ " + err.Error())
		return
	}
	dropped := m.bgSessions[idx]
	m.bgSessions = append(m.bgSessions[:idx], m.bgSessions[idx+1:]...)
	idShort := dropped.ID
	if len(idShort) > 8 {
		idShort = idShort[:8]
	}
	m.appendSystem(fmt.Sprintf("✓ 已丢弃后台 session #%d [%s]", idx+1, idShort))
}

func parseBgIndex(s string, max int) (int, error) {
	var n int
	if _, err := fmt.Sscanf(s, "%d", &n); err != nil || n < 1 || n > max {
		return 0, fmt.Errorf("无效序号 %q（1-%d）", s, max)
	}
	return n - 1, nil
}
