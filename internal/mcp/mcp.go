package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/opentmd/opentmd/internal/config"
	"github.com/opentmd/opentmd/internal/llm"
)

type ServerConfig struct {
	Command   string            `json:"command"`
	Args      []string          `json:"args"`
	Env       map[string]string `json:"env"`
	URL       string            `json:"url"`
	Disabled  bool              `json:"disabled"`
	TimeoutMS int               `json:"timeout_ms"`
}

type Config struct {
	MCPServers map[string]ServerConfig `json:"mcpServers"`
	Servers    map[string]ServerConfig `json:"servers"`
}

type ToolInfo struct {
	Server      string
	Name        string
	Description string
	Schema      map[string]any
}

type Registry struct {
	mu    sync.Mutex
	tools map[string]ToolInfo
}

func LoadConfig(workDir string) (Config, error) {
	cfg := Config{MCPServers: map[string]ServerConfig{}}
	paths := []string{}
	if dir, err := config.Dir(); err == nil {
		paths = append(paths, filepath.Join(dir, "mcp.json"))
	}
	paths = append(paths,
		filepath.Join(workDir, ".mcp.json"),
		filepath.Join(workDir, ".opentmd", "mcp.json"),
	)
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		var c Config
		if err := json.Unmarshal(data, &c); err != nil {
			return cfg, err
		}
		for k, v := range c.MCPServers {
			cfg.MCPServers[k] = v
		}
		for k, v := range c.Servers {
			if _, ok := cfg.MCPServers[k]; !ok {
				cfg.MCPServers[k] = v
			}
		}
	}
	return cfg, nil
}

func Connect(ctx context.Context, workDir string) (*Registry, error) {
	cfg, err := LoadConfig(workDir)
	if err != nil {
		return nil, err
	}
	reg := &Registry{tools: map[string]ToolInfo{}}
	for name, sc := range cfg.MCPServers {
		if sc.Disabled || sc.Command == "" {
			continue
		}
		client, err := newStdioClient(sc.Command, sc.Args, sc.Env, workDir)
		if err != nil {
			continue
		}
		tools, err := client.listTools(ctx)
		if err != nil {
			_ = client.Close()
			continue
		}
		for _, t := range tools {
			key := ToolKey(name, t.Name)
			reg.tools[key] = ToolInfo{
				Server: name, Name: t.Name, Description: t.Description, Schema: t.InputSchema,
			}
		}
		_ = client.Close()
	}
	return reg, nil
}

func ToolKey(server, tool string) string {
	return fmt.Sprintf("mcp__%s__%s", server, tool)
}

func (r *Registry) Definitions() []llm.ToolDefinition {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []llm.ToolDefinition
	for key, t := range r.tools {
		params := t.Schema
		if params == nil {
			params = map[string]any{"type": "object", "properties": map[string]any{}}
		}
		out = append(out, llm.ToolDefinition{
			Name:        key,
			Description: fmt.Sprintf("[MCP:%s] %s", t.Server, t.Description),
			Parameters:  params,
		})
	}
	return out
}

func (r *Registry) Execute(ctx context.Context, workDir, toolKey string, args json.RawMessage) (string, error) {
	r.mu.Lock()
	info, ok := r.tools[toolKey]
	r.mu.Unlock()
	if !ok {
		return "", fmt.Errorf("unknown MCP tool %q", toolKey)
	}
	cfg, _ := LoadConfig(workDir)
	sc, ok := cfg.MCPServers[info.Server]
	if !ok {
		return "", fmt.Errorf("MCP server %q not configured", info.Server)
	}
	client, err := newStdioClient(sc.Command, sc.Args, sc.Env, workDir)
	if err != nil {
		return "", err
	}
	defer client.Close()
	return client.callTool(ctx, info.Name, args)
}

type mcpTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

type stdioClient struct {
	cmd    *execWrapper
	id     atomic.Int64
	mu     sync.Mutex
}

type execWrapper struct {
	stdin  *bufio.Writer
	stdout *bufio.Scanner
	close  func() error
}

func newStdioClient(command string, args []string, env map[string]string, workDir string) (*stdioClient, error) {
	// implemented in client_stdio.go
	return dialStdio(command, args, env, workDir)
}

type jsonRPCReq struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int64  `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

type jsonRPCResp struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id"`
	Result  json.RawMessage `json:"result"`
	Error   *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (c *stdioClient) call(ctx context.Context, method string, params any) (json.RawMessage, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	id := c.id.Add(1)
	req := jsonRPCReq{JSONRPC: "2.0", ID: id, Method: method, Params: params}
	data, _ := json.Marshal(req)
	if _, err := c.cmd.stdin.Write(append(data, '\n')); err != nil {
		return nil, err
	}
	if err := c.cmd.stdin.Flush(); err != nil {
		return nil, err
	}
	if !c.cmd.stdout.Scan() {
		return nil, fmt.Errorf("mcp read: %w", c.cmd.stdout.Err())
	}
	var resp jsonRPCResp
	if err := json.Unmarshal(c.cmd.stdout.Bytes(), &resp); err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("mcp error: %s", resp.Error.Message)
	}
	return resp.Result, nil
}

func (c *stdioClient) listTools(ctx context.Context) ([]mcpTool, error) {
	_, _ = c.call(ctx, "initialize", map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]any{},
		"clientInfo":        map[string]string{"name": "opentmd", "version": "1.0"},
	})
	result, err := c.call(ctx, "tools/list", map[string]any{})
	if err != nil {
		return nil, err
	}
	var out struct {
		Tools []mcpTool `json:"tools"`
	}
	if err := json.Unmarshal(result, &out); err != nil {
		return nil, err
	}
	return out.Tools, nil
}

func (c *stdioClient) callTool(ctx context.Context, name string, args json.RawMessage) (string, error) {
	var arguments any
	if len(args) > 0 {
		_ = json.Unmarshal(args, &arguments)
	}
	result, err := c.call(ctx, "tools/call", map[string]any{"name": name, "arguments": arguments})
	if err != nil {
		return "", err
	}
	var out struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(result, &out); err != nil {
		return string(result), nil
	}
	var parts []string
	for _, c := range out.Content {
		if c.Text != "" {
			parts = append(parts, c.Text)
		}
	}
	return strings.Join(parts, "\n"), nil
}

func (c *stdioClient) Close() error {
	if c.cmd != nil && c.cmd.close != nil {
		return c.cmd.close()
	}
	return nil
}

func SaveServer(path, name string, sc ServerConfig) error {
	cfg := Config{MCPServers: map[string]ServerConfig{}}
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &cfg)
		if cfg.MCPServers == nil {
			cfg.MCPServers = map[string]ServerConfig{}
		}
	}
	cfg.MCPServers[name] = sc
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func DefaultConfigPath(global bool, workDir string) (string, error) {
	if global {
		dir, err := config.Dir()
		if err != nil {
			return "", err
		}
		return filepath.Join(dir, "mcp.json"), nil
	}
	projectPath := filepath.Join(workDir, ".opentmd", "mcp.json")
	if _, err := os.Stat(projectPath); err == nil {
		return projectPath, nil
	}
	return filepath.Join(workDir, ".mcp.json"), nil
}
