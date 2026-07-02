package permission

import "testing"

func TestClassifyMCP(t *testing.T) {
	level, reason := Classify("mcp__fs__read", "{}", "/tmp")
	if level != RequireApproval {
		t.Fatalf("expected RequireApproval, got %v", level)
	}
	if reason == "" {
		t.Fatal("expected reason")
	}
}

func TestClassifyDangerousBash(t *testing.T) {
	level, _ := Classify("bash", `{"command":"rm -rf /"}`, "/tmp")
	if level != RequireApprovalAlways {
		t.Fatalf("expected RequireApprovalAlways, got %v", level)
	}
}

func TestSessionGrant(t *testing.T) {
	s := NewStore()
	level, _ := s.Check("write_file", `{"path":"a.go"}`, "/proj")
	if level != RequireApproval {
		t.Fatalf("expected RequireApproval, got %v", level)
	}
	s.GrantSession("write_file")
	level, _ = s.Check("write_file", `{"path":"b.go"}`, "/proj")
	if level != AutoApprove {
		t.Fatalf("expected AutoApprove after grant, got %v", level)
	}
}

func TestAutoApproveRead(t *testing.T) {
	level, _ := Classify("read_file", `{"path":"main.go"}`, "/proj")
	if level != AutoApprove {
		t.Fatalf("expected AutoApprove, got %v", level)
	}
}
