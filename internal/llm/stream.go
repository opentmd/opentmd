package llm

import "sort"

type toolCallAccumulator struct {
	byIndex map[int]*ToolCall
}

func newToolCallAccumulator() *toolCallAccumulator {
	return &toolCallAccumulator{byIndex: map[int]*ToolCall{}}
}

func (a *toolCallAccumulator) add(d ToolCallDelta) {
	tc := a.byIndex[d.Index]
	if tc == nil {
		tc = &ToolCall{}
		a.byIndex[d.Index] = tc
	}
	if d.ID != "" {
		tc.ID = d.ID
	}
	if d.Name != "" {
		tc.Name = d.Name
	}
	tc.Arguments += d.Arguments
}

func (a *toolCallAccumulator) finish() []ToolCall {
	if len(a.byIndex) == 0 {
		return nil
	}
	idxs := make([]int, 0, len(a.byIndex))
	for i := range a.byIndex {
		idxs = append(idxs, i)
	}
	sort.Ints(idxs)
	out := make([]ToolCall, 0, len(idxs))
	for _, i := range idxs {
		out = append(out, *a.byIndex[i])
	}
	return out
}

// CollectStream reads a provider stream into a ChatResponse, forwarding text
// chunks to onChunk as they arrive (when no tool calls are detected yet).
func CollectStream(ch <-chan StreamChunk, onChunk func(string) error) (*ChatResponse, error) {
	var content string
	tools := newToolCallAccumulator()
	sawTool := false
	var usage Usage

	for chunk := range ch {
		if chunk.Error != nil {
			return nil, chunk.Error
		}
		if chunk.Usage.TotalTokens > 0 {
			usage = chunk.Usage
		}
		if chunk.ToolDelta != nil {
			sawTool = true
			tools.add(*chunk.ToolDelta)
		}
		if chunk.Content != "" {
			content += chunk.Content
			if onChunk != nil && !sawTool {
				if err := onChunk(chunk.Content); err != nil {
					return nil, err
				}
			}
		}
		if chunk.Done {
			break
		}
	}

	return &ChatResponse{
		Content:   content,
		ToolCalls: tools.finish(),
		Usage:     usage,
	}, nil
}
