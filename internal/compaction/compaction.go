// Package compaction compresses long session histories — ported from AtomCode's
// manual /compact flow (simplified for OpenTMD's user/assistant-only session store).
package compaction

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/opentmd/opentmd/internal/llm"
	"github.com/opentmd/opentmd/internal/session"
)

const (
	DefaultKeepRecent = 6
	MinMessages       = 8
	SummaryPrefix     = "[Previous conversation summary]\n"
)

const compressionSystemPrompt = "You are a conversation summarizer. Output ONLY the summary."

// Plan describes which messages to replace with a summary.
type Plan struct {
	RemoveCount int
	Content     string
}

// Outcome reports the result of applying compaction.
type Outcome struct {
	Applied         bool
	RemovedMessages int
	BeforeChars     int
	AfterChars      int
}

// PlanSession decides whether compaction is worthwhile.
func PlanSession(messages []session.Message, keepRecent int) (Plan, bool) {
	if keepRecent <= 0 {
		keepRecent = DefaultKeepRecent
	}
	n := len(messages)
	if n < MinMessages || n <= keepRecent+2 {
		return Plan{}, false
	}
	removeCount := n - keepRecent
	var sb strings.Builder
	for _, m := range messages[:removeCount] {
		role := string(m.Role)
		content := strings.TrimSpace(m.Content)
		if content == "" {
			continue
		}
		if len(content) > 2000 {
			content = content[:2000] + "..."
		}
		fmt.Fprintf(&sb, "%s: %s\n", role, content)
	}
	content := strings.TrimSpace(sb.String())
	if content == "" {
		return Plan{}, false
	}
	return Plan{RemoveCount: removeCount, Content: content}, true
}

// BuildSummarizePrompt creates the LLM prompt for summarization.
func BuildSummarizePrompt(content, focus string) string {
	if focus = strings.TrimSpace(focus); focus != "" {
		return fmt.Sprintf(
			"Summarize this conversation history, focusing on: %s.\n"+
				"Keep: file names, what was changed, key decisions, errors encountered.\n"+
				"Drop: exact code content, tool arguments, line numbers.\n\n%s",
			focus, content,
		)
	}
	return fmt.Sprintf(
		"Summarize this conversation history in 3-5 concise sentences. "+
			"Keep: file names, what was changed, key decisions, errors encountered. "+
			"Drop: exact code content, tool arguments, line numbers.\n\n%s",
		content,
	)
}

// Summarize calls the provider to produce a compact summary.
func Summarize(ctx context.Context, prov llm.Provider, model, prompt string) (string, error) {
	req := llm.ChatRequest{
		Model: model,
		Messages: []llm.Message{
			{Role: llm.RoleSystem, Content: compressionSystemPrompt},
			{Role: llm.RoleUser, Content: prompt},
		},
	}
	resp, err := prov.Complete(ctx, req)
	if err != nil {
		return "", err
	}
	summary := strings.TrimSpace(resp.Content)
	if summary == "" {
		return "", fmt.Errorf("empty summary from provider")
	}
	return summary, nil
}

// Apply replaces the oldest messages with a single summary user message.
func Apply(sess *session.Store, removeCount int, summary string) Outcome {
	cur := sess.Current()
	if cur == nil || removeCount <= 0 || removeCount >= len(cur.Messages) {
		return Outcome{}
	}

	beforeChars := estimateChars(cur.Messages)
	summary = strings.TrimSpace(summary)
	if summary == "" {
		return Outcome{BeforeChars: beforeChars, AfterChars: beforeChars}
	}

	kept := append([]session.Message{}, cur.Messages[removeCount:]...)
	compactMsg := session.Message{
		Role:      llm.RoleUser,
		Content:   SummaryPrefix + summary,
		Timestamp: time.Now(),
	}
	newMsgs := append([]session.Message{compactMsg}, kept...)
	afterChars := estimateChars(newMsgs)

	if afterChars >= beforeChars {
		return Outcome{BeforeChars: beforeChars, AfterChars: afterChars}
	}

	if err := sess.ReplaceMessages(newMsgs); err != nil {
		return Outcome{BeforeChars: beforeChars, AfterChars: beforeChars}
	}
	return Outcome{
		Applied:         true,
		RemovedMessages: removeCount,
		BeforeChars:     beforeChars,
		AfterChars:      afterChars,
	}
}

func estimateChars(messages []session.Message) int {
	n := 0
	for _, m := range messages {
		n += len(m.Content)
	}
	return n
}

// FormatOutcome returns a user-facing status line.
func FormatOutcome(o Outcome) string {
	if !o.Applied && o.BeforeChars == 0 && o.AfterChars == 0 {
		return "（无需压缩 — 当前对话较短）"
	}
	if !o.Applied {
		return fmt.Sprintf(
			"（压缩未生效 — 摘要未减小上下文：%s → %s）",
			fmtK(o.BeforeChars), fmtK(o.AfterChars),
		)
	}
	return fmt.Sprintf(
		"已压缩 — 合并 %d 条消息，上下文 %s → %s",
		o.RemovedMessages, fmtK(o.BeforeChars), fmtK(o.AfterChars),
	)
}

func fmtK(chars int) string {
	if chars < 1024 {
		return fmt.Sprintf("%dB", chars)
	}
	return fmt.Sprintf("%.1fK", float64(chars)/1024)
}
