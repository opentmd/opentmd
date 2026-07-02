package compaction

import (
	"testing"
	"time"

	"github.com/opentmd/opentmd-cli/internal/config"
	"github.com/opentmd/opentmd-cli/internal/llm"
	"github.com/opentmd/opentmd-cli/internal/session"
)

func msg(role llm.Role, content string) session.Message {
	return session.Message{Role: role, Content: content, Timestamp: time.Now()}
}

func TestPlanSessionTooShort(t *testing.T) {
	msgs := []session.Message{
		msg(llm.RoleUser, "a"),
		msg(llm.RoleAssistant, "b"),
	}
	if _, ok := PlanSession(msgs, 6); ok {
		t.Fatal("expected no plan for short session")
	}
}

func TestPlanSessionEnoughMessages(t *testing.T) {
	msgs := make([]session.Message, 10)
	for i := range msgs {
		if i%2 == 0 {
			msgs[i] = msg(llm.RoleUser, "question")
		} else {
			msgs[i] = msg(llm.RoleAssistant, "answer")
		}
	}
	plan, ok := PlanSession(msgs, 4)
	if !ok {
		t.Fatal("expected plan")
	}
	if plan.RemoveCount != 6 {
		t.Fatalf("remove count = %d, want 6", plan.RemoveCount)
	}
	if plan.Content == "" {
		t.Fatal("expected mechanical content")
	}
}

func TestApplyReducesSize(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	cfg := testConfig(t)
	store, err := session.NewStore(cfg)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 10; i++ {
		role := llm.RoleUser
		if i%2 == 1 {
			role = llm.RoleAssistant
		}
		if err := store.AddMessage(role, stringsRepeat("x", 500)); err != nil {
			t.Fatal(err)
		}
	}

	out := Apply(store, 6, "short summary of earlier turns")
	if !out.Applied {
		t.Fatalf("expected applied, got %#v", out)
	}
	if len(store.Current().Messages) != 5 {
		t.Fatalf("expected 5 messages after compact, got %d", len(store.Current().Messages))
	}
	if !stringsHasPrefix(store.Current().Messages[0].Content, SummaryPrefix) {
		t.Fatalf("expected summary prefix, got %q", store.Current().Messages[0].Content)
	}
}

func TestApplyNoSavingsRollsBack(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	cfg := testConfig(t)
	store, err := session.NewStore(cfg)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 10; i++ {
		role := llm.RoleUser
		if i%2 == 1 {
			role = llm.RoleAssistant
		}
		if err := store.AddMessage(role, "short"); err != nil {
			t.Fatal(err)
		}
	}
	before := len(store.Current().Messages)
	out := Apply(store, 6, stringsRepeat("y", 5000))
	if out.Applied {
		t.Fatal("expected rollback when summary is larger")
	}
	if len(store.Current().Messages) != before {
		t.Fatalf("messages changed on rollback: %d -> %d", before, len(store.Current().Messages))
	}
}

func TestBuildSummarizePromptFocus(t *testing.T) {
	p := BuildSummarizePrompt("content here", "auth changes")
	if !stringsContains(p, "auth changes") || !stringsContains(p, "content here") {
		t.Fatalf("unexpected prompt: %s", p)
	}
}

func stringsRepeat(s string, n int) string {
	out := make([]byte, 0, len(s)*n)
	for i := 0; i < n; i++ {
		out = append(out, s...)
	}
	return string(out)
}

func stringsHasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func stringsContains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func testConfig(t *testing.T) *config.Config {
	t.Helper()
	cfg := config.DefaultConfig()
	cfg.Session.Persist = true
	return cfg
}
