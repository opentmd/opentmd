package web

import (
	"net"
	"testing"
)

func TestCheckIPBlocksPrivate(t *testing.T) {
	cases := []string{"127.0.0.1", "10.0.0.1", "192.168.1.1", "169.254.169.254", "::1"}
	for _, c := range cases {
		if err := checkIP(net.ParseIP(c)); err == nil {
			t.Fatalf("expected block for %s", c)
		}
	}
}

func TestCheckIPAllowsPublic(t *testing.T) {
	if err := checkIP(net.ParseIP("8.8.8.8")); err != nil {
		t.Fatalf("expected allow: %v", err)
	}
}
