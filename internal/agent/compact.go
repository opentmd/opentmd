
package agent

import (
	"context"
	"fmt"

	"github.com/opentmd/opentmd/internal/compaction"
)

// Compact summarizes older session messages to reduce context size.
// Optional focus narrows the summary (e.g. "auth refactor").
func (a *Agent) Compact(ctx context.Context, focus string) (string, error) {
	cur := a.session.Current()
	if cur == nil || len(cur.Messages) == 0 {
		return compaction.FormatOutcome(compaction.Outcome{}), nil
	}

	plan, ok := compaction.PlanSession(cur.Messages, compaction.DefaultKeepRecent)
	if !ok {
		return compaction.FormatOutcome(compaction.Outcome{}), nil
	}

	prompt := compaction.BuildSummarizePrompt(plan.Content, focus)
	summary, err := compaction.Summarize(ctx, a.prov, a.cfg.Default.Model, prompt)
	if err != nil {
		// Fall back to mechanical summary when LLM fails.
		summary = plan.Content
	}

	out := compaction.Apply(a.session, plan.RemoveCount, summary)
	return compaction.FormatOutcome(out), nil
}

// CompactWithStatus runs compaction and reports progress via onStatus.
func (a *Agent) CompactWithStatus(ctx context.Context, focus string, onStatus StatusHandler) (string, error) {
	if onStatus != nil {
		onStatus("▸ Compacting conversation history...")
	}
	result, err := a.Compact(ctx, focus)
	if err != nil {
		return "", fmt.Errorf("compact: %w", err)
	}
	return result, nil
}
