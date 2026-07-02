package agent

import "testing"

func TestLoopGuardThirdIdenticalCallBlocked(t *testing.T) {
	g := newLoopGuard()
	args := `{"command":"cargo check"}`
	out := "error: no method named foo"

	if blocked, _ := g.check("bash", args); blocked {
		t.Fatal("first call should allow")
	}
	g.record("bash", args, out, false)

	if blocked, _ := g.check("bash", args); blocked {
		t.Fatal("second call should allow")
	}
	g.record("bash", args, out, false)

	if blocked, _ := g.check("bash", args); !blocked {
		t.Fatal("third identical call should block")
	}
}

func TestLoopGuardInterveningEditResets(t *testing.T) {
	g := newLoopGuard()
	bashArgs := `{"command":"cargo test"}`
	editArgs := `{"path":"src/lib.rs","old_string":"a","new_string":"b"}`
	testOut := "test failed"

	g.record("bash", bashArgs, testOut, false)
	g.record("bash", bashArgs, testOut, false)
	g.record("edit_file", editArgs, "ok", true)

	if blocked, _ := g.check("bash", bashArgs); blocked {
		t.Fatal("post-edit bash should be allowed")
	}
	g.record("bash", bashArgs, testOut, false)
	g.record("bash", bashArgs, testOut, false)
	if blocked, _ := g.check("bash", bashArgs); !blocked {
		t.Fatal("third post-edit bash should block")
	}
}

func TestLoopGuardHardCapWithDriftingOutput(t *testing.T) {
	g := newLoopGuard()
	args := `{"command":"cargo check 2>&1 | tail -20"}`
	for i := 0; i < 5; i++ {
		if blocked, _ := g.check("bash", args); blocked {
			t.Fatalf("call %d should allow", i+1)
		}
		g.record("bash", args, "warning drift "+string(rune('0'+i)), false)
	}
	if blocked, msg := g.check("bash", args); !blocked {
		t.Fatal("6th call should hard-cap block")
	} else if msg == "" {
		t.Fatal("expected block message")
	}
}
