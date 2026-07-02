package llm

import "context"

type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
	RoleTool      Role = "tool"
)

type ToolCall struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type ToolDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

type Message struct {
	Role       Role       `json:"role"`
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

type ChatRequest struct {
	Model    string
	Messages []Message
	Stream   bool
	Tools    []ToolDefinition
}

type StreamChunk struct {
	Content   string
	ToolDelta *ToolCallDelta
	Usage     Usage
	Done      bool
	Error     error
}

// ToolCallDelta is an incremental tool-call fragment from a streaming response.
type ToolCallDelta struct {
	Index     int
	ID        string
	Name      string
	Arguments string
}

type ChatResponse struct {
	Content   string
	ToolCalls []ToolCall
	Usage     Usage
}

type Provider interface {
	Chat(ctx context.Context, req ChatRequest) (<-chan StreamChunk, error)
	Complete(ctx context.Context, req ChatRequest) (*ChatResponse, error)
}
