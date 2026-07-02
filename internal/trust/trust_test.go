package trust

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTrustStore(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	// config.Dir() -> ~/.opentmd
	home := dir
	_ = os.MkdirAll(filepath.Join(home, ".opentmd"), 0o755)

	ws := filepath.Join(dir, "project")
	if err := os.MkdirAll(ws, 0o755); err != nil {
		t.Fatal(err)
	}

	ok, err := IsTrusted(ws)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("expected untrusted workspace")
	}

	if err := Trust(ws); err != nil {
		t.Fatal(err)
	}
	ok, err = IsTrusted(ws)
	if err != nil || !ok {
		t.Fatalf("expected trusted, ok=%v err=%v", ok, err)
	}

	if err := Untrust(ws); err != nil {
		t.Fatal(err)
	}
	ok, err = IsTrusted(ws)
	if err != nil || ok {
		t.Fatalf("expected untrusted after revoke, ok=%v err=%v", ok, err)
	}
}

func TestResolveSkipAndAutoTrust(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	_ = os.MkdirAll(filepath.Join(dir, ".opentmd"), 0o755)
	ws := filepath.Join(dir, "repo")
	_ = os.MkdirAll(ws, 0o755)

	trusted, need, err := Resolve(ws, Options{Skip: true})
	if err != nil || !trusted || need {
		t.Fatalf("skip: trusted=%v need=%v err=%v", trusted, need, err)
	}

	trusted, need, err = Resolve(ws, Options{AutoTrust: true})
	if err != nil || !trusted || need {
		t.Fatalf("auto trust: trusted=%v need=%v err=%v", trusted, need, err)
	}
	ok, _ := IsTrusted(ws)
	if !ok {
		t.Fatal("auto trust should persist")
	}
}
