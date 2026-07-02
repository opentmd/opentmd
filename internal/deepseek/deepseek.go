package deepseek

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/opentmd/opentmd/internal/llm"
)

const defaultTimeout = 120 * time.Second

type Client struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

func New(baseURL, apiKey string) *Client {
	baseURL = strings.TrimRight(baseURL, "/")
	if !strings.HasSuffix(baseURL, "/v1") {
		baseURL += "/v1"
	}
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		client:  &http.Client{Timeout: defaultTimeout},
	}
}

type chatRequestBody struct {
	Model    string           `json:"model"`
	Messages []messagePayload `json:"messages"`
	Stream   bool             `json:"stream"`
	Tools    []toolPayload    `json:"tools,omitempty"`
}

type messagePayload struct {
	Role       string             `json:"role"`
	Content    string             `json:"content"`
	ToolCalls  []toolCallPayload  `json:"tool_calls,omitempty"`
	ToolCallID string             `json:"tool_call_id,omitempty"`
}

type toolPayload struct {
	Type     string         `json:"type"`
	Function functionSchema `json:"function"`
}

type functionSchema struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

type toolCallPayload struct {
	Index    *int   `json:"index,omitempty"`
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type streamResponse struct {
	Usage   llm.Usage `json:"usage"`
	Choices []struct {
		Delta struct {
			Content   string            `json:"content"`
			ToolCalls []toolCallPayload `json:"tool_calls"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}

type completeResponse struct {
	Usage   llm.Usage `json:"usage"`
	Choices []struct {
		Message struct {
			Role      string            `json:"role"`
			Content   string            `json:"content"`
			ToolCalls []toolCallPayload `json:"tool_calls"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

func (c *Client) Chat(ctx context.Context, req llm.ChatRequest) (<-chan llm.StreamChunk, error) {
	body, err := c.buildBody(req, true)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	c.setHeaders(httpReq, true)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}

	ch := make(chan llm.StreamChunk, 64)
	go c.readStream(ctx, resp.Body, ch)
	return ch, nil
}

func (c *Client) Complete(ctx context.Context, req llm.ChatRequest) (*llm.ChatResponse, error) {
	return c.CompleteWithRetry(ctx, req, DefaultRetryPolicy())
}

func (c *Client) completeOnce(ctx context.Context, req llm.ChatRequest) (*llm.ChatResponse, error) {
	body, err := c.buildBody(req, false)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	c.setHeaders(httpReq, false)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		body := strings.TrimSpace(string(b))
		return nil, &apiStatusError{
			Code:       resp.StatusCode,
			Body:       body,
			RetryAfter: parseRetryAfter(resp.Header),
		}
	}

	var cr completeResponse
	if err := json.Unmarshal(b, &cr); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	if len(cr.Choices) == 0 {
		return nil, fmt.Errorf("empty response from API")
	}

	msg := cr.Choices[0].Message
	out := &llm.ChatResponse{Content: msg.Content, Usage: cr.Usage}
	for _, tc := range msg.ToolCalls {
		out.ToolCalls = append(out.ToolCalls, llm.ToolCall{
			ID:        tc.ID,
			Name:      tc.Function.Name,
			Arguments: tc.Function.Arguments,
		})
	}
	return out, nil
}

func (c *Client) buildBody(req llm.ChatRequest, stream bool) ([]byte, error) {
	messages := make([]messagePayload, len(req.Messages))
	for i, m := range req.Messages {
		mp := messagePayload{
			Role:       string(m.Role),
			Content:    m.Content,
			ToolCallID: m.ToolCallID,
		}
		for _, tc := range m.ToolCalls {
			mp.ToolCalls = append(mp.ToolCalls, toolCallPayload{
				ID:   tc.ID,
				Type: "function",
				Function: struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				}{Name: tc.Name, Arguments: tc.Arguments},
			})
		}
		messages[i] = mp
	}

	payload := chatRequestBody{
		Model:    req.Model,
		Messages: messages,
		Stream:   stream,
	}
	for _, t := range req.Tools {
		payload.Tools = append(payload.Tools, toolPayload{
			Type: "function",
			Function: functionSchema{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.Parameters,
			},
		})
	}
	return json.Marshal(payload)
}

func (c *Client) setHeaders(req *http.Request, stream bool) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	if stream {
		req.Header.Set("Accept", "text/event-stream")
	}
}

func (c *Client) readStream(ctx context.Context, body io.ReadCloser, ch chan<- llm.StreamChunk) {
	defer close(ch)
	defer body.Close()

	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	for {
		if ctx.Err() != nil {
			ch <- llm.StreamChunk{Error: ctx.Err()}
			return
		}

		if !scanner.Scan() {
			break
		}

		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			ch <- llm.StreamChunk{Done: true}
			return
		}

		var sr streamResponse
		if err := json.Unmarshal([]byte(data), &sr); err != nil {
			ch <- llm.StreamChunk{Error: fmt.Errorf("parse stream chunk: %w", err)}
			return
		}
		if len(sr.Choices) == 0 {
			continue
		}
		choice := sr.Choices[0]
		if sr.Usage.TotalTokens > 0 {
			ch <- llm.StreamChunk{Usage: sr.Usage}
		}
		if content := choice.Delta.Content; content != "" {
			ch <- llm.StreamChunk{Content: content}
		}
		for _, tc := range choice.Delta.ToolCalls {
			idx := 0
			if tc.Index != nil {
				idx = *tc.Index
			}
			ch <- llm.StreamChunk{ToolDelta: &llm.ToolCallDelta{
				Index:     idx,
				ID:        tc.ID,
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			}}
		}
		if choice.FinishReason != nil && *choice.FinishReason != "" {
			ch <- llm.StreamChunk{Done: true}
			return
		}
	}

	if err := scanner.Err(); err != nil {
		ch <- llm.StreamChunk{Error: fmt.Errorf("read stream: %w", err)}
	}
}

// ParseSSEData parses one SSE data line (exported for tests).
func ParseSSEData(data string) (content string, done bool, err error) {
	if data == "[DONE]" {
		return "", true, nil
	}
	var sr streamResponse
	if err := json.Unmarshal([]byte(data), &sr); err != nil {
		return "", false, err
	}
	if len(sr.Choices) == 0 {
		return "", false, nil
	}
	content = sr.Choices[0].Delta.Content
	if sr.Choices[0].FinishReason != nil && *sr.Choices[0].FinishReason != "" {
		done = true
	}
	return content, done, nil
}
