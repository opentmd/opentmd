package agent

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/opentmd/opentmd/internal/config"
	"github.com/opentmd/opentmd/internal/hook"
	"github.com/opentmd/opentmd/internal/llm"
	"github.com/opentmd/opentmd/internal/lsp"
	"github.com/opentmd/opentmd/internal/mcp"
	"github.com/opentmd/opentmd/internal/memory"
	"github.com/opentmd/opentmd/internal/pathutil"
	"github.com/opentmd/opentmd/internal/permission"
	"github.com/opentmd/opentmd/internal/project"
	"github.com/opentmd/opentmd/internal/provider"
	"github.com/opentmd/opentmd/internal/session"
	"github.com/opentmd/opentmd/internal/tool"
)

type Options struct {
	WorkDir     string
	Verbose     bool
	AutoApprove bool
	OnApproval  permission.Handler
}

type Agent struct {
	cfg     *config.Config
	prov    llm.Provider
	session *session.Store
	tools   *tool.Runtime
	hooks   *hook.Executor
	opts    Options
	usage   *UsageStats

	memoryGlobal  *memory.Store
	memoryProject *memory.Store
}

func New(cfg *config.Config, sess *session.Store, opts Options) (*Agent, error) {
	prov, err := provider.New(cfg)
	if err != nil {
		return nil, err
	}
	workDir := opts.WorkDir
	if workDir == "" {
		workDir, _ = os.Getwd()
	}
	rt, err := tool.NewRuntime(workDir, tool.Options{
		AutoApprove: opts.AutoApprove,
		OnApproval:  opts.OnApproval,
		Config:      cfg,
	})
	if err != nil {
		return nil, err
	}
	hooks, _ := hook.Load(workDir)
	ag := &Agent{
		cfg:     cfg,
		prov:    prov,
		session: sess,
		tools:   rt,
		hooks:   hooks,
		opts:    opts,
		usage:   &UsageStats{},

		memoryGlobal:  memory.Global(),
		memoryProject: memory.Project(workDir),
	}
	ag.ensureFileHistory()
	return ag, nil
}

func (a *Agent) SetApprovalHandler(h permission.Handler) {
	a.opts.OnApproval = h
	a.tools.SetApprovalHandler(h)
}

type StreamHandler func(chunk string) error
type StatusHandler func(status string)

func (a *Agent) Chat(ctx context.Context, userInput string, onChunk StreamHandler) (string, error) {
	return a.ChatWithStatus(ctx, userInput, onChunk, nil)
}

func (a *Agent) ChatWithStatus(ctx context.Context, userInput string, onChunk StreamHandler, onStatus StatusHandler) (string, error) {
	userInput = strings.TrimSpace(userInput)
	if userInput == "" {
		return "", nil
	}

	if a.hooks != nil {
		var ok bool
		userInput, ok = a.hooks.RunUserPrompt(ctx, userInput)
		if !ok {
			return "", fmt.Errorf("prompt blocked by hook")
		}
	}

	history := a.session.ProviderMessages()
	messages := append([]llm.Message{}, history...)
	messages = append(messages, llm.Message{Role: llm.RoleUser, Content: userInput})

	a.tools.BeginTurn()
	response, err := a.runLoop(ctx, messages, onChunk, onStatus)
	if err != nil {
		return "", err
	}

	if err := a.session.AddMessage(llm.RoleUser, userInput); err != nil {
		return response, fmt.Errorf("save user message: %w", err)
	}
	if err := a.session.AddMessage(llm.RoleAssistant, response); err != nil {
		return response, fmt.Errorf("save assistant message: %w", err)
	}
	return response, nil
}

func (a *Agent) runLoop(ctx context.Context, messages []llm.Message, onChunk StreamHandler, onStatus StatusHandler) (string, error) {
	memoryPrompt := a.memoryPrompt()
	system := buildSystemPrompt(a.tools.WorkDir(), a.tools.SkillsPrompt(), memoryPrompt, a.tools.PlanMode())
	full := append([]llm.Message{{Role: llm.RoleSystem, Content: system}}, messages...)
	maxIter := a.cfg.Agent.MaxIterations
	if maxIter <= 0 {
		maxIter = 15
	}

	guard := newLoopGuard()

	for i := 0; i < maxIter; i++ {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}

		resp, err := a.completeTurn(ctx, full, onChunk)
		if err != nil {
			return "", err
		}
		if a.usage != nil {
			a.usage.Add(resp.Usage)
		}

		if len(resp.ToolCalls) > 0 {
			full = append(full, llm.Message{
				Role:      llm.RoleAssistant,
				Content:   resp.Content,
				ToolCalls: resp.ToolCalls,
			})
			for _, tc := range resp.ToolCalls {
				status := describeTool(tc.Name, tc.Arguments)
				if onStatus != nil {
					onStatus(status)
				}
				if a.opts.Verbose {
					fmt.Fprintln(os.Stderr, status)
				}

				var result string
				var execErr error
				if blocked, msg := guard.check(tc.Name, tc.Arguments); blocked {
					result = msg
					execErr = fmt.Errorf("loop guard")
				} else {
					result, execErr = a.tools.Execute(ctx, tc)
					if execErr != nil {
						result = "error: " + execErr.Error()
					}
					success := execErr == nil && !strings.HasPrefix(result, "error:")
					guard.record(tc.Name, tc.Arguments, result, success)
				}

				full = append(full, llm.Message{
					Role:       llm.RoleTool,
					Content:    result,
					ToolCallID: tc.ID,
				})
			}
			continue
		}

		content := resp.Content
		return content, nil
	}
	return "", fmt.Errorf("agent exceeded max iterations (%d)", maxIter)
}

func (a *Agent) completeTurn(ctx context.Context, messages []llm.Message, onChunk StreamHandler) (*llm.ChatResponse, error) {
	req := llm.ChatRequest{
		Model:    a.cfg.Default.Model,
		Messages: messages,
		Tools:    a.tools.Definitions(),
	}
	if onChunk != nil && a.cfg.TUI.Stream {
		ch, err := a.prov.Chat(ctx, req)
		if err != nil {
			return nil, err
		}
		return llm.CollectStream(ch, func(chunk string) error {
			return onChunk(chunk)
		})
	}
	return a.prov.Complete(ctx, req)
}

func (a *Agent) ChatToWriter(ctx context.Context, userInput string, w io.Writer) (string, error) {
	return a.Chat(ctx, userInput, func(chunk string) error {
		_, err := io.WriteString(w, chunk)
		return err
	})
}

func (a *Agent) Reload(cfg *config.Config) error {
	prov, err := provider.New(cfg)
	if err != nil {
		return err
	}
	a.cfg = cfg
	a.prov = prov
	a.tools.ReloadLSP(cfg)
	return nil
}

func (a *Agent) SetLSPEventHandler(h lsp.EventHandler) {
	a.tools.SetLSPEventHandler(h)
}

func (a *Agent) Shutdown() {
	if a.tools != nil {
		a.tools.Close()
	}
}

func (a *Agent) ReloadMCP(ctx context.Context) error {
	return a.tools.ReloadMCP(ctx)
}

func (a *Agent) ReloadLSP() {
	a.tools.ReloadLSP(a.cfg)
}

func (a *Agent) LSPStatus() string {
	return a.tools.LSPStatus()
}

func (a *Agent) LSPStatusStruct() lsp.Status {
	return a.tools.LSPStatusStruct()
}

func (a *Agent) MCPStatus(ctx context.Context) mcp.StatusReport {
	return a.tools.MCPStatus(ctx)
}

func (a *Agent) ExpandSkill(name, args string) (string, error) {
	return a.tools.ExpandSkill(name, args)
}

func (a *Agent) ListSkills() []string {
	var names []string
	for _, s := range a.tools.ListSkills() {
		names = append(names, s.Name)
	}
	return names
}

func (a *Agent) SetWorkDir(dir string) error {
	abs, err := pathutil.Resolve(a.tools.WorkDir(), dir)
	if err != nil {
		return fmt.Errorf("resolve work dir: %w", err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return fmt.Errorf("work dir: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", abs)
	}
	if err := project.InitWorkspace(abs, a.cfg); err != nil {
		return err
	}
	a.tools.SetWorkDir(abs)
	a.memoryProject = memory.Project(abs)
	a.hooks, _ = hook.Load(abs)
	return os.Chdir(abs)
}

func (a *Agent) WorkDir() string         { return a.tools.WorkDir() }
func (a *Agent) Config() *config.Config  { return a.cfg }
func (a *Agent) Session() *session.Store { return a.session }
func (a *Agent) SetPlanMode(enabled bool) {
	a.tools.SetPlanMode(enabled)
}
func (a *Agent) PlanMode() bool {
	return a.tools.PlanMode()
}
func (a *Agent) ClearSession() error {
	a.ResetUsage()
	if err := a.session.Clear(); err != nil {
		return err
	}
	a.ensureFileHistory()
	return nil
}

func (a *Agent) LoadSession(id string) error {
	if err := a.session.Load(id); err != nil {
		return err
	}
	a.ensureFileHistory()
	return nil
}
func (a *Agent) ListSessions() ([]session.Session, error) { return a.session.List() }
func (a *Agent) UndoLastTurn() ([]string, error) {
	return a.tools.UndoLastTurn()
}

func (a *Agent) ensureFileHistory() {
	if s := a.session.Current(); s != nil {
		_ = a.tools.InitFileHistory(s.ID)
	}
}

func (a *Agent) NewSession() error {
	if err := a.session.New(); err != nil {
		return err
	}
	a.ensureFileHistory()
	return nil
}

// memoryPrompt returns merged memory entries as a system-prompt block.
func (a *Agent) memoryPrompt() string {
	return memory.MergedForPrompt(a.memoryGlobal, a.memoryProject, a.projectName())
}

// Remember appends an entry to the appropriate memory store.
func (a *Agent) Remember(scope, content string) error {
	content = strings.TrimSpace(content)
	if content == "" {
		return fmt.Errorf("memory content is empty")
	}
	switch scope {
	case "global":
		return a.memoryGlobal.Append(content)
	case "project":
		return a.memoryProject.Append(content)
	default:
		// Default to global
		return a.memoryGlobal.Append(content)
	}
}

// Forget removes entries matching the keyword from the given scope.
func (a *Agent) Forget(scope, keyword string) ([]string, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return nil, fmt.Errorf("keyword is empty")
	}
	switch scope {
	case "project":
		return a.memoryProject.RemoveMatching(keyword)
	default:
		return a.memoryGlobal.RemoveMatching(keyword)
	}
}

// MemoryList returns all memory entries for the given scope.
func (a *Agent) MemoryList(scope string) []string {
	switch scope {
	case "project":
		return a.memoryProject.Load()
	case "all":
		global := a.memoryGlobal.Load()
		project := a.memoryProject.Load()
		all := make([]string, 0, len(global)+len(project)+2)
		if len(global) > 0 {
			all = append(all, "── Global ──")
			all = append(all, global...)
		}
		if len(project) > 0 {
			all = append(all, "── Project ──")
			all = append(all, project...)
		}
		return all
	default:
		return a.memoryGlobal.Load()
	}
}

func (a *Agent) projectName() string {
	wd := a.WorkDir()
	return filepath.Base(wd)
}
