package lsp

import (
	"strings"
	"testing"
	"time"
)

func TestFormatStatusDisabled(t *testing.T) {
	out := FormatStatus(Status{Enabled: false})
	if !strings.Contains(out, "disabled") {
		t.Fatalf("expected disabled message, got %q", out)
	}
}

func TestFormatStatusActiveAndIdle(t *testing.T) {
	out := FormatStatus(Status{
		Enabled:         true,
		AutoDetect:      true,
		SettleDelayMs:   300,
		WarmupFileLimit: 40,
		Servers: []ServerEntry{
			{Ext: "go", Command: "gopls", Active: true, InPath: true},
			{Ext: "rs", Command: "rust-analyzer", Active: false, InPath: true},
			{Ext: "java", Command: "jdtls", Active: false, InPath: false},
		},
	})
	for _, want := range []string{"已启动", "gopls", "未启动", "rust-analyzer", "未安装", "jdtls"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in output, got %q", want, out)
		}
	}
}

func TestManagerStatusUsesRegistry(t *testing.T) {
	mgr := NewManager(t.TempDir(), Config{
		Enabled:         true,
		AutoDetect:      true,
		SettleDelay:     100 * time.Millisecond,
		WarmupFileLimit: 5,
	}, nil)
	st := mgr.Status()
	if !st.Enabled {
		t.Fatal("expected enabled")
	}
	if len(st.Servers) == 0 {
		t.Fatal("expected configured servers from auto_detect defaults")
	}
}
