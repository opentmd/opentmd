package permission

import (
	"encoding/json"
	"path/filepath"
	"strings"
)

type Level int

const (
	AutoApprove Level = iota
	RequireApproval
	RequireApprovalAlways
)

type Request struct {
	Tool   string
	Args   string
	Reason string
	Level  Level
}

type Decision int

const (
	Deny Decision = iota
	AllowOnce
	AllowSession
)

type Handler func(req Request) Decision

type Store struct {
	sessionGrants map[string]struct{}
}

func NewStore() *Store {
	return &Store{sessionGrants: map[string]struct{}{}}
}

func (s *Store) GrantSession(tool string) {
	s.sessionGrants[tool] = struct{}{}
}

func (s *Store) Allowed(tool string) bool {
	_, ok := s.sessionGrants[tool]
	return ok
}

func Classify(tool, args, workDir string) (Level, string) {
	if strings.HasPrefix(tool, "mcp__") {
		return RequireApproval, "external MCP tool"
	}
	switch tool {
	case "write_file", "edit_file", "search_replace":
		if outsideWorkspace(args, workDir, "path") {
			return RequireApprovalAlways, "write outside workspace"
		}
		return RequireApproval, "file modification"
	case "bash", "run_shell":
		cmd := extractField(args, "command")
		if isDangerousShell(cmd) {
			return RequireApprovalAlways, "dangerous shell command"
		}
		return RequireApproval, "shell execution"
	default:
		return AutoApprove, ""
	}
}

func (s *Store) Check(tool, args, workDir string) (Level, string) {
	level, reason := Classify(tool, args, workDir)
	if level == RequireApprovalAlways {
		return level, reason
	}
	if level == RequireApproval && s.Allowed(tool) {
		return AutoApprove, ""
	}
	return level, reason
}

func outsideWorkspace(argsJSON, workDir, field string) bool {
	val := extractField(argsJSON, field)
	if val == "" {
		return false
	}
	if filepath.IsAbs(val) {
		rel, err := filepath.Rel(workDir, val)
		return err != nil || strings.HasPrefix(rel, "..")
	}
	return false
}

func extractField(argsJSON, field string) string {
	var m map[string]any
	if err := json.Unmarshal([]byte(argsJSON), &m); err != nil {
		return ""
	}
	v, _ := m[field].(string)
	return v
}

func isDangerousShell(cmd string) bool {
	cmd = strings.ToLower(strings.TrimSpace(cmd))
	patterns := []string{
		"rm -rf", "rm -fr", "sudo ", "mkfs", "dd if=", ":(){", "chmod -r",
		"> /dev/sd", "shutdown", "reboot", "curl | sh", "wget | sh",
	}
	for _, p := range patterns {
		if strings.Contains(cmd, p) {
			return true
		}
	}
	return false
}

func AutoHandler() Handler {
	return func(req Request) Decision { return AllowOnce }
}

func DenyHandler() Handler {
	return func(req Request) Decision { return Deny }
}
