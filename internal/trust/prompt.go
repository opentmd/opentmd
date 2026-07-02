package trust

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// EnsureTerminal prompts on a TTY when the workspace is not yet trusted.
func EnsureTerminal(workDir string, opts Options) error {
	trusted, needPrompt, err := Resolve(workDir, opts)
	if err != nil {
		return err
	}
	if trusted || !needPrompt {
		return nil
	}
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return fmt.Errorf("workspace not trusted: %s (run in a terminal, or use --trust)", workDir)
	}
	return promptTerminal(workDir)
}

func promptTerminal(workDir string) error {
	fmt.Fprint(os.Stderr, renderPromptBox(workDir, 0))
	fmt.Fprint(os.Stderr, "\n  按 a 信任此工作区，q 退出\n")

	reader := bufio.NewReader(os.Stdin)
	for {
		key, err := readKey(reader)
		if err != nil {
			return err
		}
		switch key {
		case "enter", "a":
			fmt.Fprintln(os.Stderr)
			return Trust(workDir)
		case "q":
			fmt.Fprintln(os.Stderr)
			return ErrDeclined
		}
	}
}

func renderPromptBox(workDir string, selected int) string {
	const width = 72
	border := strings.Repeat("─", width-2)
	lines := []string{
		"  ╭" + border + "╮",
		"  │" + padLine("  ⚠ Workspace Trust Required", width-2) + "│",
		"  │" + padLine("", width-2) + "│",
		"  │" + padLine("  OpenTMD can execute code and access files in this directory.", width-2) + "│",
		"  │" + padLine("", width-2) + "│",
		"  │" + padLine("  Do you trust the contents of this directory?", width-2) + "│",
		"  │" + padLine("", width-2) + "│",
		"  │" + padLine("    "+workDir, width-2) + "│",
		"  │" + padLine("", width-2) + "│",
	}
	choices := []string{
		"  ▶ [a] Trust this workspace",
		"    [q] Quit",
	}
	if selected == 1 {
		choices[0] = "    [a] Trust this workspace"
		choices[1] = "  ▶ [q] Quit"
	}
	for _, c := range choices {
		lines = append(lines, "  │"+padLine(c, width-2)+"│")
	}
	lines = append(lines,
		"  │"+padLine("", width-2)+"│",
		"  │"+padLine("  Use arrow keys to navigate, Enter to select, or press the key shown", width-2)+"│",
		"  │"+padLine("", width-2)+"│",
		"  ╰"+border+"╯",
	)
	return strings.Join(lines, "\n") + "\n"
}

func padLine(s string, width int) string {
	if len(s) >= width {
		if width <= 1 {
			return s[:width]
		}
		return s[:width-1] + "…"
	}
	return s + strings.Repeat(" ", width-len(s))
}

func readKey(r *bufio.Reader) (string, error) {
	b, err := r.ReadByte()
	if err != nil {
		return "", err
	}
	switch b {
	case '\n', '\r':
		return "enter", nil
	case 'a', 'A':
		return "a", nil
	case 'q', 'Q':
		return "q", nil
	case 3, 4, 27:
		if b == 27 {
			seq, _ := r.Peek(2)
			if len(seq) == 2 && seq[0] == '[' {
				switch seq[1] {
				case 'A':
					_, _ = r.ReadByte()
					_, _ = r.ReadByte()
					return "up", nil
				case 'B':
					_, _ = r.ReadByte()
					_, _ = r.ReadByte()
					return "down", nil
				}
			}
		}
		return "q", nil
	case 'j':
		return "down", nil
	case 'k':
		return "up", nil
	default:
		return string(b), nil
	}
}
