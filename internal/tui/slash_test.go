package tui

import "testing"

func TestFilterSlashCommands(t *testing.T) {
	all := filterSlashCommands("/")
	if len(all) != len(slashCommands) {
		t.Fatalf("expected all commands for /, got %d", len(all))
	}

	help := filterSlashCommands("/h")
	if len(help) != 1 || help[0].Name != "/help" {
		t.Fatalf("expected /help, got %#v", help)
	}

	session := filterSlashCommands("/session")
	if len(session) != 1 || session[0].Name != "/session" {
		t.Fatalf("expected /session, got %#v", session)
	}

	none := filterSlashCommands("/zzzz")
	if len(none) != 0 {
		t.Fatalf("expected no matches, got %#v", none)
	}
}

func TestSlashMenuWindow(t *testing.T) {
	start, end, scroll := slashMenuWindow(0, 12, 9, 0)
	if start != 0 || end != 9 || scroll != 0 {
		t.Fatalf("unexpected window: %d %d %d", start, end, scroll)
	}

	start, end, scroll = slashMenuWindow(10, 12, 9, 0)
	if start != 2 || end != 11 || scroll != 2 {
		t.Fatalf("unexpected tail window: %d %d %d", start, end, scroll)
	}
}

func TestClampSlashSelection(t *testing.T) {
	if got := clampSlashSelection(5, 3); got != 2 {
		t.Fatalf("expected 2, got %d", got)
	}
	if got := clampSlashSelection(-1, 3); got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
}

func TestCycleSlashSelection(t *testing.T) {
	if got := cycleSlashSelection(0, 5, true); got != 4 {
		t.Fatalf("up at top should wrap to last, got %d", got)
	}
	if got := cycleSlashSelection(4, 5, false); got != 0 {
		t.Fatalf("down at bottom should wrap to first, got %d", got)
	}
	if got := cycleSlashSelection(2, 5, true); got != 1 {
		t.Fatalf("expected 1, got %d", got)
	}
	if got := cycleSlashSelection(2, 5, false); got != 3 {
		t.Fatalf("expected 3, got %d", got)
	}
}
