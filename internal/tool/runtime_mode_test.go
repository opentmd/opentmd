package tool

import "testing"

func TestPlanModeBlockedTool(t *testing.T) {
	cases := []struct {
		name    string
		blocked bool
	}{
		{name: "write_file", blocked: true},
		{name: "edit_file", blocked: true},
		{name: "search_replace", blocked: true},
		{name: "bash", blocked: true},
		{name: "mcp__github__search", blocked: true},
		{name: "read_file", blocked: false},
		{name: "grep", blocked: false},
		{name: "todo", blocked: false},
	}
	for _, tc := range cases {
		if got := planModeBlockedTool(tc.name); got != tc.blocked {
			t.Fatalf("tool %s blocked=%v, got %v", tc.name, tc.blocked, got)
		}
	}
}

func TestDefinitionsFilterInPlanMode(t *testing.T) {
	rt := &Runtime{
		Registry: *NewRegistry(t.TempDir()),
		planMode: true,
	}
	defs := rt.Definitions()
	names := map[string]struct{}{}
	for _, d := range defs {
		names[d.Name] = struct{}{}
	}
	if _, ok := names["write_file"]; ok {
		t.Fatalf("write_file should be hidden in plan mode")
	}
	if _, ok := names["bash"]; ok {
		t.Fatalf("bash should be hidden in plan mode")
	}
	if _, ok := names["read_file"]; !ok {
		t.Fatalf("read_file should remain available in plan mode")
	}
}
