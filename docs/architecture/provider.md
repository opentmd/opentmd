# Provider 系统

## 概述

Provider 系统位于 `internal/provider/` + `internal/deepseek/` + `internal/llm/`，定义了统一的 LLM 调用接口和 OpenAI 兼容的 HTTP 客户端实现。

## LLM 接口 (`internal/llm/`)

### 类型定义

```go
type Role string

const (
    RoleUser      Role = "user"
    RoleAssistant Role = "assistant"
    RoleSystem    Role = "system"
    RoleTool      Role = "tool"
)

type ToolCall struct {
    ID        string
    Name      string
    Arguments string
}

type ToolDefinition struct {
    Name        string
    Description string
    Parameters  map[string]any
}

type Message struct {
    Role         Role
    Content      string
    ToolCalls    []ToolCall
    ToolCallID   string
}

type ChatRequest struct {
    Model    string
    Messages []Message
    Tools    []ToolDefinition
    Stream   bool
}

type ChatResponse struct {
    Content   string
    ToolCalls []ToolCall
}

type StreamChunk struct {
    Content string
    Done    bool
    Error   error
}
```

### Provider 接口

```go
type Provider interface {
    Chat(ctx context.Context, req ChatRequest) (<-chan StreamChunk, error)
    Complete(ctx context.Context, req ChatRequest) (*ChatResponse, error)
}
```

- `Chat()` — 流式调用，返回 channel
- `Complete()` — 非流式调用，返回完整响应

## Provider 工厂 (`internal/provider/`)

根据配置选择对应 Provider，目前所有 Provider 统一使用 `deepseek.Client`（OpenAI 兼容 API）。

```go
func New(cfg *config.Config) (llm.Provider, error) {
    name, pc, err := cfg.ActiveProvider()
    // 统一使用 deepseek.New(baseURL, apiKey)
}
```

## DeepSeek 客户端 (`internal/deepseek/`)

### 特点
- OpenAI 兼容 HTTP 客户端
- 支持流式 (SSE) 和非流式调用
- 支持 Tool Calling (function calling)
- 超时默认 120 秒

### 关键结构

```go
type Client struct {
    baseURL string
    apiKey  string
    client  *http.Client
}
```

自动拼接 `/v1/chat/completions` 端点。

### 请求体

```json
{
  "model": "deepseek-chat",
  "messages": [...],
  "stream": true,
  "tools": [...]  // 可选
}
```

### 流式响应解析

SSE 格式：
```
data: {"choices":[{"delta":{"content":"Hello"},"finish_reason":null}]}
data: [DONE]
```

工具调用通过 `delta.tool_calls` 字段传递。

### 测试辅助

`ParseSSEData()` 函数导出用于单元测试，解析单条 SSE `data:` 行内容。
