package memory

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAppendCreatesFile(t *testing.T) {
	dir := t.TempDir()
	store := New(filepath.Join(dir, "sub", "memory.md"))
	if err := store.Append("test entry"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(store.path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "- test entry\n" {
		t.Fatalf("got %q, want %q", string(data), "- test entry\n")
	}
}

func TestAppendToExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "memory.md")
	if err := os.WriteFile(path, []byte("- first\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	store := New(path)
	if err := store.Append("second"); err != nil {
		t.Fatal(err)
	}
	entries := store.Load()
	if len(entries) != 2 || entries[0] != "first" || entries[1] != "second" {
		t.Fatalf("got %v, want [first second]", entries)
	}
}

func TestLoadSkipsNonEntries(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "memory.md")
	if err := os.WriteFile(path, []byte("# Header\n\n- real entry\nnot an entry\n- another\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	store := New(path)
	entries := store.Load()
	if len(entries) != 2 || entries[0] != "real entry" || entries[1] != "another" {
		t.Fatalf("got %v, want [real entry another]", entries)
	}
}

func TestLoadEmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "memory.md")
	if err := os.WriteFile(path, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	store := New(path)
	if entries := store.Load(); len(entries) != 0 {
		t.Fatalf("got %v, want empty", entries)
	}
}

func TestLoadNonexistent(t *testing.T) {
	store := New("/nonexistent/memory.md")
	if entries := store.Load(); len(entries) != 0 {
		t.Fatalf("got %v, want empty", entries)
	}
}

func TestRemoveMatchingCaseInsensitive(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "memory.md")
	if err := os.WriteFile(path, []byte("- Use tabs\n- use spaces\n- pnpm only\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	store := New(path)
	removed, err := store.RemoveMatching("use")
	if err != nil {
		t.Fatal(err)
	}
	if len(removed) != 2 || removed[0] != "Use tabs" || removed[1] != "use spaces" {
		t.Fatalf("got %v, want [Use tabs use spaces]", removed)
	}
	entries := store.Load()
	if len(entries) != 1 || entries[0] != "pnpm only" {
		t.Fatalf("got %v, want [pnpm only]", entries)
	}
}

func TestRemoveMatchingNoMatch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "memory.md")
	if err := os.WriteFile(path, []byte("- keep this\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	store := New(path)
	removed, err := store.RemoveMatching("nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	if len(removed) != 0 {
		t.Fatalf("got %v, want empty", removed)
	}
	data, _ := os.ReadFile(path)
	if string(data) != "- keep this\n" {
		t.Fatalf("file was modified: %q", string(data))
	}
}

func TestMergedForPromptEmpty(t *testing.T) {
	dir := t.TempDir()
	global := New(filepath.Join(dir, "global.md"))
	project := New(filepath.Join(dir, "proj", ".opentmd", "memory.md"))
	result := MergedForPrompt(global, project, "testproj")
	if result != "" {
		t.Fatalf("expected empty, got %q", result)
	}
}

func TestMergedForPromptWithEntries(t *testing.T) {
	dir := t.TempDir()
	global := New(filepath.Join(dir, "global.md"))
	project := New(filepath.Join(dir, "project.md"))
	if err := global.Append("prefers tabs"); err != nil {
		t.Fatal(err)
	}
	if err := project.Append("pnpm only"); err != nil {
		t.Fatal(err)
	}
	result := MergedForPrompt(global, project, "myproj")
	if !contains(result, "prefers tabs") {
		t.Fatalf("result missing 'prefers tabs': %s", result)
	}
	if !contains(result, "pnpm only") {
		t.Fatalf("result missing 'pnpm only': %s", result)
	}
	if !contains(result, "[Project: myproj]") {
		t.Fatalf("result missing project label: %s", result)
	}
}

func TestFindMatching(t *testing.T) {
	dir := t.TempDir()
	store := New(filepath.Join(dir, "memory.md"))
	if err := store.Append("use tabs for indentation"); err != nil {
		t.Fatal(err)
	}
	if err := store.Append("prefer spaces"); err != nil {
		t.Fatal(err)
	}
	if err := store.Append("pnpm for packages"); err != nil {
		t.Fatal(err)
	}

	matches := store.FindMatching("tabs")
	if len(matches) != 1 || matches[0] != "use tabs for indentation" {
		t.Fatalf("got %v, want [use tabs for indentation]", matches)
	}

	matches = store.FindMatching("TABS")
	if len(matches) != 1 {
		t.Fatalf("case-insensitive search failed: %v", matches)
	}
}

func TestValidInvalid(t *testing.T) {
	store := New("")
	if store.Valid() {
		t.Fatal("expected invalid store")
	}
	store = New("/tmp/test.md")
	if !store.Valid() {
		t.Fatal("expected valid store")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsStr(s, substr)
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
