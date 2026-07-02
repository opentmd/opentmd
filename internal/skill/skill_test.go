package skill

import "testing"

func TestParseSkillFrontmatter(t *testing.T) {
	raw := `---
name: my-skill
description: "Do something"
---

Body with $ARGUMENTS here`
	s := parseSkill(raw, "/tmp/my-skill")
	if s.Name != "my-skill" {
		t.Fatalf("name = %q", s.Name)
	}
	if s.Description != "Do something" {
		t.Fatalf("desc = %q", s.Description)
	}
	if !contains(s.Content, "$ARGUMENTS") {
		t.Fatal("body missing placeholder")
	}
}

func TestExpandSkill(t *testing.T) {
	r := &Registry{skills: map[string]Skill{
		"test": {Name: "test", Content: "Hello $ARGUMENTS"},
	}}
	out, err := r.Expand("test", "world")
	if err != nil || out != "Hello world" {
		t.Fatalf("expand = %q err=%v", out, err)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
