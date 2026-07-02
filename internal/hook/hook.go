package hook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/opentmd/opentmd/internal/config"
)

type Event string

const (
	PreToolUse        Event = "PreToolUse"
	PostToolUse       Event = "PostToolUse"
	UserPromptSubmit  Event = "UserPromptSubmit"
	SessionStart      Event = "SessionStart"
)

type Entry struct {
	Event     Event  `json:"event"`
	Matcher   string `json:"matcher"`
	Command   string `json:"command"`
	TimeoutMS int    `json:"timeout_ms"`
	Disabled  bool   `json:"disabled"`
}

type Config struct {
	Hooks map[string]Entry `json:"hooks"`
}

type PreResult struct {
	Allow  bool
	Reason string
	Args   string
}

type Executor struct {
	entries []Entry
}

func Load(workDir string) (*Executor, error) {
	var all []Entry
	if cfgDir, err := config.Dir(); err == nil {
		if e, err := loadFile(filepath.Join(cfgDir, "hooks.json")); err == nil {
			all = append(all, e...)
		}
	}
	for _, name := range []string{".hooks.json", filepath.Join(".opentmd", "hooks.json")} {
		if e, err := loadFile(filepath.Join(workDir, name)); err == nil {
			all = append(all, e...)
		}
	}
	return &Executor{entries: all}, nil
}

func loadFile(path string) ([]Entry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	var out []Entry
	for _, e := range cfg.Hooks {
		if !e.Disabled && e.Command != "" {
			out = append(out, e)
		}
	}
	return out, nil
}

func (e *Executor) RunPre(ctx context.Context, tool, args string) PreResult {
	for _, h := range e.entries {
		if h.Event != PreToolUse || !match(h.Matcher, tool) {
			continue
		}
		out, err := e.run(ctx, h, map[string]string{
			"OPENTMD_HOOK_EVENT":  string(PreToolUse),
			"OPENTMD_TOOL_NAME":   tool,
			"OPENTMD_HOOK_CONTEXT": args,
		}, "")
		if err != nil {
			continue
		}
		var resp struct {
			Action string `json:"action"`
			Reason string `json:"reason"`
			Args   string `json:"args"`
		}
		if json.Unmarshal([]byte(out), &resp) == nil {
			switch strings.ToLower(resp.Action) {
			case "block", "deny":
				return PreResult{Allow: false, Reason: resp.Reason}
			case "modify":
				if resp.Args != "" {
					return PreResult{Allow: true, Args: resp.Args}
				}
			}
		}
	}
	return PreResult{Allow: true, Args: args}
}

func (e *Executor) RunPost(ctx context.Context, tool, args, result string) {
	for _, h := range e.entries {
		if h.Event != PostToolUse || !match(h.Matcher, tool) {
			continue
		}
		_, _ = e.run(ctx, h, map[string]string{
			"OPENTMD_HOOK_EVENT": string(PostToolUse),
			"OPENTMD_TOOL_NAME":  tool,
		}, result)
	}
}

func (e *Executor) RunUserPrompt(ctx context.Context, prompt string) (string, bool) {
	for _, h := range e.entries {
		if h.Event != UserPromptSubmit {
			continue
		}
		payload, _ := json.Marshal(map[string]string{"prompt": prompt})
		out, err := e.run(ctx, h, map[string]string{
			"OPENTMD_HOOK_EVENT": string(UserPromptSubmit),
		}, string(payload))
		if err != nil {
			continue
		}
		var resp struct {
			Action  string `json:"action"`
			Prompt  string `json:"prompt"`
			Reason  string `json:"reason"`
		}
		if json.Unmarshal([]byte(out), &resp) == nil && strings.ToLower(resp.Action) == "block" {
			return resp.Reason, false
		}
		if resp.Prompt != "" {
			return resp.Prompt, true
		}
	}
	return prompt, true
}

func match(matcher, tool string) bool {
	if matcher == "" || matcher == "*" {
		return true
	}
	if strings.HasSuffix(matcher, "*") {
		return strings.HasPrefix(tool, strings.TrimSuffix(matcher, "*"))
	}
	return matcher == tool
}

func (e *Executor) run(ctx context.Context, h Entry, env map[string]string, stdin string) (string, error) {
	timeout := time.Duration(h.TimeoutMS) * time.Millisecond
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", h.Command)
	cmd.Env = os.Environ()
	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return strings.TrimSpace(buf.String()), nil
}
