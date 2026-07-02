package tui

import "testing"

func TestChunkMsgTargetsAssistantAfterToolStatus(t *testing.T) {
	m := Model{
		messages: []displayMsg{
			{kind: "user", text: "check system info"},
			{kind: "assistant", text: ""},
			{kind: "tool", text: "▸ Running {\"command\": \"uname -a\"}"},
		},
	}

	idx := m.lastAssistantIdx()
	if idx != 1 {
		t.Fatalf("expected assistant idx 1, got %d", idx)
	}

	if idx := m.lastAssistantIdx(); idx >= 0 {
		m.messages[idx].text += "Hello"
	}
	if m.messages[1].text != "Hello" {
		t.Fatalf("expected chunk on assistant message, got %q", m.messages[1].text)
	}
}

func TestSetAssistantContentFillsEmptyReply(t *testing.T) {
	m := Model{
		messages: []displayMsg{
			{kind: "user", text: "hi"},
			{kind: "assistant", text: ""},
			{kind: "tool", text: "▸ Running ls"},
		},
	}

	m.setAssistantContent("done")
	if m.messages[1].text != "done" {
		t.Fatalf("expected assistant content %q, got %q", "done", m.messages[1].text)
	}
}

func TestSetAssistantContentKeepsStreamedText(t *testing.T) {
	m := Model{
		messages: []displayMsg{
			{kind: "assistant", text: "partial"},
		},
	}

	m.setAssistantContent("final")
	if m.messages[0].text != "partial" {
		t.Fatalf("expected streamed text preserved, got %q", m.messages[0].text)
	}
}
