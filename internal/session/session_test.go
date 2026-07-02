package session_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/opentmd/opentmd-cli/internal/config"
	"github.com/opentmd/opentmd-cli/internal/llm"
	"github.com/opentmd/opentmd-cli/internal/session"
)

func testConfig(t *testing.T) *config.Config {
	t.Helper()
	cfg := config.DefaultConfig()
	cfg.Session.Persist = true
	return cfg
}

func TestSessionRoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	store, err := session.NewStoreWithOptions(testConfig(t), session.Options{ContinueLast: true})
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	if err := store.AddMessage(llm.RoleUser, "hello"); err != nil {
		t.Fatalf("add message: %v", err)
	}
	if err := store.AddMessage(llm.RoleAssistant, "world"); err != nil {
		t.Fatalf("add message: %v", err)
	}

	id := store.Current().ID
	store2, err := session.NewStoreWithOptions(testConfig(t), session.Options{ContinueLast: true})
	if err != nil {
		t.Fatalf("reopen store: %v", err)
	}
	if store2.Current().ID != id {
		t.Fatalf("expected restored session %s, got %s", id, store2.Current().ID)
	}
	if len(store2.Current().Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(store2.Current().Messages))
	}
}

func TestSessionDefaultStartsFresh(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	first, err := session.NewStoreWithOptions(testConfig(t), session.Options{ContinueLast: true})
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	if err := first.AddMessage(llm.RoleUser, "persist me"); err != nil {
		t.Fatal(err)
	}
	oldID := first.Current().ID

	fresh, err := session.NewStore(testConfig(t))
	if err != nil {
		t.Fatalf("fresh store: %v", err)
	}
	if fresh.Current().ID == oldID {
		t.Fatalf("expected new session id, got same %s", oldID)
	}
	if len(fresh.Current().Messages) != 0 {
		t.Fatalf("expected empty session, got %d messages", len(fresh.Current().Messages))
	}
}

func TestSessionLoadRejectsPathTraversal(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	store, err := session.NewStore(testConfig(t))
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	if err := store.Load("../etc/passwd"); err == nil {
		t.Fatal("expected error for path traversal id")
	}
}

func TestTruncateTitleUnicode(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	store, err := session.NewStore(testConfig(t))
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	long := strings.Repeat("你好世界", 15)
	if err := store.AddMessage(llm.RoleUser, long); err != nil {
		t.Fatal(err)
	}
	title := store.Current().Title
	if title == "" {
		t.Fatal("title should not be empty")
	}
	if filepath.Base(title) == title && len([]rune(title)) > 43 {
		t.Fatalf("title not truncated properly: %q", title)
	}
}
