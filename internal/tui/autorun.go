package tui

import (
	"fmt"
	"strings"

	"github.com/opentmd/opentmd/internal/permission"
)

func (m *Model) handleAutoRun(args []string) {
	if len(args) == 0 {
		m.autoRun = !m.autoRun
		m.appendAutoRunStatus(m.autoRun, true)
		return
	}
	switch strings.ToLower(args[0]) {
	case "on", "enable", "true", "1":
		m.autoRun = true
		m.appendAutoRunStatus(true, false)
	case "off", "disable", "false", "0":
		m.autoRun = false
		m.appendAutoRunStatus(false, false)
	case "status":
		m.appendAutoRunStatus(m.autoRun, false)
	default:
		m.appendSystem("用法: /auto-run [on|off|status]")
	}
}

func (m *Model) appendAutoRunStatus(on bool, toggled bool) {
	state := "关闭"
	if on {
		state = "开启"
	}
	msg := fmt.Sprintf("Auto-run: %s", state)
	if toggled {
		msg = "✓ 已切换 — " + msg
	}
	msg += "（常规工具自动执行；危险操作仍会确认）"
	m.appendSystem(msg)
}

func (m *Model) resolveApproval(req permission.Request) permission.Decision {
	if m.autoRun && req.Level != permission.RequireApprovalAlways {
		return permission.AllowOnce
	}
	if m.program != nil {
		m.program.Send(approvalMsg{req: req})
	}
	if m.approvalCh == nil {
		return permission.Deny
	}
	return <-m.approvalCh
}
