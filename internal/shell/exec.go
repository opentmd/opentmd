package shell

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const DefaultTimeout = 60 * time.Second
const DefaultMaxOutput = 32 * 1024

type Result struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func Run(ctx context.Context, command string, workDir string, timeout time.Duration, maxOutput int) (*Result, error) {
	command = strings.TrimSpace(command)
	if command == "" {
		return nil, fmt.Errorf("command is empty")
	}
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	if maxOutput <= 0 {
		maxOutput = DefaultMaxOutput
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", "-c", command)
	if workDir != "" {
		cmd.Dir = workDir
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	res := &Result{
		Stdout: truncate(stdout.String(), maxOutput),
		Stderr: truncate(stderr.String(), maxOutput),
	}
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			res.ExitCode = exitErr.ExitCode()
			return res, nil
		}
		if ctx.Err() != nil {
			return res, ctx.Err()
		}
		return res, fmt.Errorf("run command: %w", err)
	}
	return res, nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "\n[output truncated]"
}

func FormatResult(res *Result) string {
	var sb strings.Builder
	if res.Stdout != "" {
		sb.WriteString("stdout:\n")
		sb.WriteString(res.Stdout)
	}
	if res.Stderr != "" {
		if sb.Len() > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString("stderr:\n")
		sb.WriteString(res.Stderr)
	}
	sb.WriteString(fmt.Sprintf("\nexit_code: %d", res.ExitCode))
	return sb.String()
}
