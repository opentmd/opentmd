package tui

import (
	"testing"

	"github.com/opentmd/opentmd/internal/permission"
)

func TestResolveApprovalAutoRunSkipsPrompt(t *testing.T) {
	ch := make(chan permission.Decision, 1)
	m := &Model{
		autoRun:    true,
		approvalCh: ch,
	}
	dec := m.resolveApproval(permission.Request{
		Tool: "bash", Level: permission.RequireApproval, Reason: "shell execution",
	})
	if dec != permission.AllowOnce {
		t.Fatalf("expected AllowOnce, got %v", dec)
	}
	select {
	case <-ch:
		t.Fatal("should not send approval prompt when auto-run is on")
	default:
	}
}

func TestResolveApprovalAutoRunStillPromptsDangerous(t *testing.T) {
	ch := make(chan permission.Decision, 1)
	m := &Model{
		autoRun:    true,
		approvalCh: ch,
	}
	go func() {
		ch <- permission.Deny
	}()
	dec := m.resolveApproval(permission.Request{
		Tool: "bash", Level: permission.RequireApprovalAlways, Reason: "dangerous shell command",
	})
	if dec != permission.Deny {
		t.Fatalf("expected Deny, got %v", dec)
	}
}
