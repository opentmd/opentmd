package deepseek_test

import (
	"testing"

	"github.com/opentmd/opentmd-cli/internal/deepseek"
)

func TestParseSSEDataContent(t *testing.T) {
	data := `{"choices":[{"delta":{"content":"hello"},"finish_reason":null}]}`
	content, done, err := deepseek.ParseSSEData(data)
	if err != nil {
		t.Fatal(err)
	}
	if content != "hello" {
		t.Fatalf("expected hello, got %q", content)
	}
	if done {
		t.Fatal("expected not done")
	}
}

func TestParseSSEDataDone(t *testing.T) {
	content, done, err := deepseek.ParseSSEData("[DONE]")
	if err != nil {
		t.Fatal(err)
	}
	if content != "" || !done {
		t.Fatalf("expected done, got content=%q done=%v", content, done)
	}
}

func TestParseSSEDataFinishReason(t *testing.T) {
	data := `{"choices":[{"delta":{},"finish_reason":"stop"}]}`
	_, done, err := deepseek.ParseSSEData(data)
	if err != nil {
		t.Fatal(err)
	}
	if !done {
		t.Fatal("expected done when finish_reason is set")
	}
}
