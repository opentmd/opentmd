package mcp

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
)

func dialStdio(command string, args []string, env map[string]string, workDir string) (*stdioClient, error) {
	cmd := exec.Command(command, args...)
	cmd.Dir = workDir
	cmd.Env = os.Environ()
	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	w := bufio.NewWriter(stdin)
	s := bufio.NewScanner(stdout)
	s.Buffer(make([]byte, 64*1024), 1024*1024)
	return &stdioClient{
		cmd: &execWrapper{
			stdin:  w,
			stdout: s,
			close: func() error {
				_ = stdin.Close()
				return cmd.Process.Kill()
			},
		},
	}, nil
}
