package agent

import (
	"fmt"
	"strings"

	"github.com/opentmd/opentmd/internal/project"
)

func buildSystemPrompt(workDir, skillsPrompt string, memoryPrompt string, planMode bool) string {
	var sb strings.Builder
	sb.WriteString(baseSystemPrompt)
	sb.WriteString("\n\nWorking directory: ")
	sb.WriteString(workDir)
	if planMode {
		sb.WriteString("\n\nMode: PLAN (read-only)")
		sb.WriteString("\n- Do not modify files.")
		sb.WriteString("\n- Do not execute shell commands.")
		sb.WriteString("\n- Focus on investigation, analysis, and implementation planning.")
	}

	if stack := project.DetectStack(workDir); stack != "" {
		sb.WriteString("\n")
		sb.WriteString(stack)
	}

	if ctx := project.LoadContext(workDir); ctx != "" {
		sb.WriteString("\n\n")
		sb.WriteString(ctx)
	}
	if skillsPrompt != "" {
		sb.WriteString("\n\n")
		sb.WriteString(skillsPrompt)
	}
	if memoryPrompt != "" {
		sb.WriteString("\n\n")
		sb.WriteString(memoryPrompt)
	}
	return sb.String()
}

const baseSystemPrompt = `You are OpenTMD, an AI coding assistant running in the user's terminal.

You can autonomously explore and modify code using tools. Work in a loop until the task is done:
read/grep/glob → edit/write → bash (test/build) → verify → respond.

Tools:
- read_file, list_directory, glob, grep — explore the codebase
- list_symbols, read_symbol — multi-language symbol index (Go/Rust/Python/JS/TS/Java/C/C++)
- edit_file, write_file, search_replace — modify code
- bash — run tests, builds, linters
- diagnostics — LSP real-time diagnostics (gopls, rust-analyzer, tsserver, pylsp)
- web_search, web_fetch — search and fetch public URLs
- use_skill — load skill templates
- todo — track multi-step task progress
- trace_callers, trace_callees, file_dependencies — Go code graph (when indexed)
- mcp__* — external MCP tools when configured

Guidelines:
- Use grep/glob before editing when unsure where code lives.
- Prefer edit_file for single-file changes; search_replace for bulk renames.
- Always verify with bash after edits when tests exist.
- Keep responses concise; summarize what you changed.
- Never fabricate file contents or command output.`

func describeTool(name, args string) string {
	args = strings.TrimSpace(args)
	if len(args) > 80 {
		args = args[:80] + "..."
	}
	switch name {
	case "read_file", "grep", "glob", "list_directory":
		return fmt.Sprintf("▸ Reading %s(%s)", name, args)
	case "edit_file", "write_file", "search_replace":
		return fmt.Sprintf("▸ Editing %s(%s)", name, args)
	case "bash", "run_shell":
		return fmt.Sprintf("▸ Running %s", truncateArg(args, 60))
	default:
		return fmt.Sprintf("▸ %s(%s)", name, args)
	}
}

func truncateArg(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
