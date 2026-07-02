package daemon

import (
	"testing"

	"github.com/opentmd/opentmd-cli/internal/permission"
)

func TestParseDecision(t *testing.T) {
	tests := map[string]permission.Decision{
		"allow_once":     permission.AllowOnce,
		"y":              permission.AllowOnce,
		"allow_session":  permission.AllowSession,
		"a":              permission.AllowSession,
		"deny":           permission.Deny,
		"n":              permission.Deny,
		"unknown":        permission.Deny,
	}
	for in, want := range tests {
		if got := parseDecision(in); got != want {
			t.Fatalf("parseDecision(%q) = %v want %v", in, got, want)
		}
	}
}

func TestPermissionDelivery(t *testing.T) {
	s := &Server{permPending: map[string]chan permission.Decision{}}
	ch := s.registerPermission("req-1")
	go func() {
		s.deliverPermission("req-1", permission.AllowOnce)
	}()
	if d := <-ch; d != permission.AllowOnce {
		t.Fatalf("got %v", d)
	}
	if s.deliverPermission("missing", permission.AllowOnce) {
		t.Fatal("expected false for missing id")
	}
}
