package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/opentmd/opentmd/internal/config"
	"github.com/opentmd/opentmd/internal/diagnostics"
	"github.com/opentmd/opentmd/internal/filehistory"
	"github.com/opentmd/opentmd/internal/graph"
	"github.com/opentmd/opentmd/internal/hook"
	"github.com/opentmd/opentmd/internal/llm"
	"github.com/opentmd/opentmd/internal/lsp"
	"github.com/opentmd/opentmd/internal/mcp"
	"github.com/opentmd/opentmd/internal/permission"
	"github.com/opentmd/opentmd/internal/skill"
)

type Options struct {
	AutoApprove     bool
	OnApproval      permission.Handler
	Config          *config.Config
	LSPEventHandler lsp.EventHandler
}

type Runtime struct {
	Registry
	cfg        *config.Config
	perm       *permission.Store
	hooks      *hook.Executor
	skills     *skill.Registry
	mcp        *mcp.Registry
	graph      *graph.Index
	todos      *todoList
	lsp        *diagnostics.Collector
	lspHandler lsp.EventHandler
	fileHist   *filehistory.Session
	opts       Options
	planMode   bool
}

func NewRuntime(workDir string, opts Options) (*Runtime, error) {
	if opts.OnApproval == nil && !opts.AutoApprove {
		opts.OnApproval = func(req permission.Request) permission.Decision {
			return permission.AllowOnce
		}
	}
	if opts.AutoApprove {
		opts.OnApproval = permission.AutoHandler()
	}

	hooks, _ := hook.Load(workDir)
	skills, _ := skill.Load(workDir)
	mcpReg, _ := mcp.Connect(context.Background(), workDir)
	gidx := graph.NewIndex(workDir)
	_ = gidx.LoadOrBuild()

	r := &Runtime{
		Registry:   *NewRegistry(workDir),
		cfg:        opts.Config,
		perm:       permission.NewStore(),
		hooks:      hooks,
		skills:     skills,
		mcp:        mcpReg,
		graph:      gidx,
		todos:      newTodoList(),
		lspHandler: opts.LSPEventHandler,
		opts:       opts,
	}
	r.initLSP(workDir)
	r.registerExtended()
	return r, nil
}

func (r *Runtime) registerExtended() {
	r.registerTodo()
	r.registerWebTools()
	r.registerSemanticTools()
	r.registerDiagnosticsTool()
	if r.skills != nil {
		r.Register(llm.ToolDefinition{
			Name:        "use_skill",
			Description: "Load a skill template by name and return expanded instructions.",
			Parameters: objParam(map[string]any{
				"name":      strParam("Skill name"),
				"arguments": strParam("Arguments for the skill template"),
			}, "name"),
		}, r.useSkill)
	}
	if r.graph != nil {
		r.Register(llm.ToolDefinition{
			Name:        "trace_callers",
			Description: "Find callers of a symbol in the Go code graph.",
			Parameters: objParam(map[string]any{
				"symbol": strParam("Function or symbol name"),
			}, "symbol"),
		}, r.traceCallers)
		r.Register(llm.ToolDefinition{
			Name:        "trace_callees",
			Description: "Trace callees of a symbol in the Go code graph (forward call chain).",
			Parameters: objParam(map[string]any{
				"symbol": strParam("Function or symbol name"),
				"depth":  intParam("Max traversal depth (default 3, max 5)"),
			}, "symbol"),
		}, r.traceCallees)
		r.Register(llm.ToolDefinition{
			Name:        "file_dependencies",
			Description: "List symbols defined in a Go source file.",
			Parameters: objParam(map[string]any{
				"path": strParam("Go file path"),
			}, "path"),
		}, r.fileDeps)
	}
	if r.mcp != nil {
		for _, def := range r.mcp.Definitions() {
			name := def.Name
			r.Register(def, func(ctx context.Context, args json.RawMessage) (string, error) {
				return r.mcp.Execute(ctx, r.WorkDir(), name, args)
			})
		}
	}
}

func (r *Runtime) InitFileHistory(sessionID string) error {
	fh, err := filehistory.NewSession(sessionID)
	if err != nil {
		return err
	}
	r.fileHist = fh
	r.Registry.SetBeforeWrite(fh.Backup)
	return nil
}

func (r *Runtime) BeginTurn() {
	if r.fileHist != nil {
		r.fileHist.BeginTurn()
	}
}

func (r *Runtime) UndoLastTurn() ([]string, error) {
	if r.fileHist == nil {
		return nil, fmt.Errorf("file history not initialized")
	}
	return r.fileHist.UndoLastTurn()
}

func (r *Runtime) Definitions() []llm.ToolDefinition {
	defs := r.Registry.Definitions()
	if !r.planMode {
		return defs
	}
	out := make([]llm.ToolDefinition, 0, len(defs))
	for _, def := range defs {
		if planModeBlockedTool(def.Name) {
			continue
		}
		out = append(out, def)
	}
	return out
}

func (r *Runtime) Execute(ctx context.Context, call llm.ToolCall) (string, error) {
	if r.planMode && planModeBlockedTool(call.Name) {
		return "", fmt.Errorf("%s is disabled in /plan mode; switch to /build to enable execution", call.Name)
	}
	repaired := RepairToolArgs(call.Name, call.Arguments)
	call.Arguments = repaired

	if r.hooks != nil {
		pre := r.hooks.RunPre(ctx, call.Name, call.Arguments)
		if !pre.Allow {
			return "", fmt.Errorf("blocked by hook: %s", pre.Reason)
		}
		if pre.Args != "" {
			call.Arguments = pre.Args
		}
	}

	level, reason := r.perm.Check(call.Name, string(call.Arguments), r.WorkDir())
	if level != permission.AutoApprove {
		if r.opts.OnApproval == nil {
			return "", fmt.Errorf("permission required for %s: %s", call.Name, reason)
		}
		decision := r.opts.OnApproval(permission.Request{
			Tool: call.Name, Args: string(call.Arguments), Reason: reason, Level: level,
		})
		switch decision {
		case permission.Deny:
			return "", fmt.Errorf("permission denied for %s", call.Name)
		case permission.AllowSession:
			r.perm.GrantSession(call.Name)
		}
	}

	result, err := r.Registry.Execute(ctx, call)
	if err != nil {
		result = "error: " + err.Error()
	} else {
		r.notifyLSPAfterTool(ctx, call.Name, call.Arguments)
	}
	if r.hooks != nil {
		r.hooks.RunPost(ctx, call.Name, call.Arguments, result)
	}
	return result, err
}

func (r *Runtime) SkillsPrompt() string {
	if r.skills == nil {
		return ""
	}
	return r.skills.PromptSection()
}

func (r *Runtime) ExpandSkill(name, args string) (string, error) {
	if r.skills == nil {
		return "", fmt.Errorf("skills not loaded")
	}
	return r.skills.Expand(name, args)
}

func (r *Runtime) ListSkills() []skill.Skill {
	if r.skills == nil {
		return nil
	}
	return r.skills.List()
}

func (r *Runtime) SetApprovalHandler(h permission.Handler) {
	r.opts.OnApproval = h
}

func (r *Runtime) SetLSPEventHandler(h lsp.EventHandler) {
	r.lspHandler = h
	if r.lsp != nil {
		r.lsp.SetEventHandler(h)
	}
}

func (r *Runtime) initLSP(workDir string) {
	if r.lsp != nil {
		r.lsp.Close()
	}
	r.lsp = diagnostics.NewCollector(workDir, r.cfg, r.lspHandler)
}

func (r *Runtime) ReloadLSP(cfg *config.Config) {
	r.cfg = cfg
	r.initLSP(r.WorkDir())
}

func (r *Runtime) LSPStatus() string {
	return r.LSPStatusStruct().Text
}

func (r *Runtime) LSPStatusStruct() lsp.Status {
	if r.lsp == nil {
		return lsp.Status{Text: "LSP: not initialized"}
	}
	return r.lsp.Status()
}

func (r *Runtime) MCPStatus(ctx context.Context) mcp.StatusReport {
	return mcp.ProbeStatus(ctx, r.WorkDir())
}

func (r *Runtime) Close() {
	if r.lsp != nil {
		r.lsp.Close()
		r.lsp = nil
	}
}

func (r *Runtime) SetWorkDir(dir string) {
	r.Registry.SetWorkDir(dir)
	r.initLSP(dir)
}

func (r *Runtime) SetPlanMode(enabled bool) {
	r.planMode = enabled
}

func (r *Runtime) PlanMode() bool {
	return r.planMode
}

func planModeBlockedTool(name string) bool {
	switch {
	case strings.HasPrefix(name, "mcp__"):
		return true
	}
	switch name {
	case "write_file", "edit_file", "search_replace", "bash", "run_shell":
		return true
	default:
		return false
	}
}

func (r *Runtime) ReloadMCP(ctx context.Context) error {
	reg, err := mcp.Connect(ctx, r.WorkDir())
	if err != nil {
		return err
	}
	r.mcp = reg
	for _, def := range reg.Definitions() {
		name := def.Name
		r.Register(def, func(ctx context.Context, args json.RawMessage) (string, error) {
			return r.mcp.Execute(ctx, r.WorkDir(), name, args)
		})
	}
	return nil
}

type useSkillArgs struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

func (r *Runtime) useSkill(ctx context.Context, args json.RawMessage) (string, error) {
	a, err := parseArgs[useSkillArgs](args, "use_skill")
	if err != nil {
		return "", err
	}
	return r.skills.Expand(a.Name, a.Arguments)
}

type symbolArgs struct {
	Symbol string `json:"symbol"`
}

func (r *Runtime) traceCallers(ctx context.Context, args json.RawMessage) (string, error) {
	a, err := parseArgs[symbolArgs](args, "trace_callers")
	if err != nil {
		return "", err
	}
	return r.graph.TraceCallers(a.Symbol), nil
}

type traceCalleesArgs struct {
	Symbol string `json:"symbol"`
	Depth  int    `json:"depth"`
}

func (r *Runtime) traceCallees(ctx context.Context, args json.RawMessage) (string, error) {
	a, err := parseArgs[traceCalleesArgs](args, "trace_callees")
	if err != nil {
		return "", err
	}
	return r.graph.TraceCallees(a.Symbol, a.Depth), nil
}

type fileDepsArgs struct {
	Path string `json:"path"`
}

func (r *Runtime) fileDeps(ctx context.Context, args json.RawMessage) (string, error) {
	a, err := parseArgs[fileDepsArgs](args, "file_dependencies")
	if err != nil {
		return "", err
	}
	return r.graph.FileDependencies(a.Path), nil
}
