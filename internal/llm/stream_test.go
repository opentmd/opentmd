package llm

import "testing"

func TestToolCallAccumulatorMergesDeltas(t *testing.T) {
	acc := newToolCallAccumulator()
	acc.add(ToolCallDelta{Index: 0, ID: "call_1", Name: "read_file"})
	acc.add(ToolCallDelta{Index: 0, Arguments: `{"path":`})
	acc.add(ToolCallDelta{Index: 0, Arguments: `"a.go"}`})

	calls := acc.finish()
	if len(calls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(calls))
	}
	if calls[0].ID != "call_1" || calls[0].Name != "read_file" {
		t.Fatalf("unexpected tool call: %+v", calls[0])
	}
	if calls[0].Arguments != `{"path":"a.go"}` {
		t.Fatalf("unexpected args: %q", calls[0].Arguments)
	}
}

func TestCollectStreamForwardsContent(t *testing.T) {
	ch := make(chan StreamChunk, 4)
	go func() {
		ch <- StreamChunk{Content: "Hel"}
		ch <- StreamChunk{Content: "lo"}
		ch <- StreamChunk{Done: true}
		close(ch)
	}()

	var parts []string
	resp, err := CollectStream(ch, func(s string) error {
		parts = append(parts, s)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Content != "Hello" {
		t.Fatalf("content = %q", resp.Content)
	}
	if len(parts) != 2 || parts[0]+parts[1] != "Hello" {
		t.Fatalf("chunks = %v", parts)
	}
}

func TestCollectStreamStopsForwardingAfterToolDelta(t *testing.T) {
	ch := make(chan StreamChunk, 4)
	go func() {
		ch <- StreamChunk{Content: "wait"}
		ch <- StreamChunk{ToolDelta: &ToolCallDelta{Index: 0, Name: "grep"}}
		ch <- StreamChunk{Content: "ignored"}
		ch <- StreamChunk{Done: true}
		close(ch)
	}()

	var parts []string
	resp, err := CollectStream(ch, func(s string) error {
		parts = append(parts, s)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Content != "waitignored" {
		t.Fatalf("content = %q", resp.Content)
	}
	if len(parts) != 1 || parts[0] != "wait" {
		t.Fatalf("expected only pre-tool chunk, got %v", parts)
	}
	if len(resp.ToolCalls) != 1 || resp.ToolCalls[0].Name != "grep" {
		t.Fatalf("tool calls = %+v", resp.ToolCalls)
	}
}
