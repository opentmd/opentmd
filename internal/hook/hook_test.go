package hook

import "testing"

func TestMatch(t *testing.T) {
	tests := []struct {
		matcher, tool string
		want          bool
	}{
		{"*", "bash", true},
		{"", "bash", true},
		{"bash", "bash", true},
		{"bash", "read_file", false},
		{"mcp__*", "mcp__fs__read", true},
		{"mcp__*", "bash", false},
	}
	for _, tc := range tests {
		if got := match(tc.matcher, tc.tool); got != tc.want {
			t.Fatalf("match(%q,%q) = %v want %v", tc.matcher, tc.tool, got, tc.want)
		}
	}
}
