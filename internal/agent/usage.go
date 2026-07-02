package agent

import "github.com/opentmd/opentmd/internal/llm"

type UsageStats struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

func (s *UsageStats) Add(u llm.Usage) {
	s.PromptTokens += u.PromptTokens
	s.CompletionTokens += u.CompletionTokens
	s.TotalTokens += u.TotalTokens
}

func (s UsageStats) Snapshot() UsageStats { return s }

func (a *Agent) Usage() UsageStats {
	if a.usage == nil {
		return UsageStats{}
	}
	return a.usage.Snapshot()
}

func (a *Agent) ResetUsage() {
	a.usage = &UsageStats{}
}
