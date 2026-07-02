package tui

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/aymanbagabas/go-osc52/v2"
	"github.com/opentmd/opentmd-cli/internal/config"
)

func copyFallbackPath() (string, error) {
	dir, err := config.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "last-copy.txt"), nil
}

func writeCopyFallback(text string) (string, error) {
	path, err := copyFallbackPath()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", fmt.Errorf("create fallback dir: %w", err)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		return "", fmt.Errorf("write fallback file: %w", err)
	}
	return path, nil
}

func copyToSystemClipboard(text string) error {
	if strings.TrimSpace(text) == "" {
		return fmt.Errorf("empty text")
	}
	var lastErr error
	for _, fn := range []func(string) error{
		copyViaWlCopy,
		copyViaXclip,
		copyViaXsel,
		copyViaAtotto,
		copyViaOSC52,
	} {
		if err := fn(text); err == nil {
			return nil
		} else {
			lastErr = err
		}
	}
	if lastErr != nil {
		return lastErr
	}
	return fmt.Errorf("no clipboard backend available")
}

func copyViaCommand(name string, args []string, text string) error {
	if _, err := exec.LookPath(name); err != nil {
		return err
	}
	cmd := exec.Command(name, args...)
	cmd.Stdin = strings.NewReader(text)
	cmd.Env = os.Environ()
	if out, err := cmd.CombinedOutput(); err != nil {
		msg := strings.TrimSpace(string(out))
		if msg != "" {
			return fmt.Errorf("%s: %s: %w", name, msg, err)
		}
		return fmt.Errorf("%s: %w", name, err)
	}
	return nil
}

func copyViaWlCopy(text string) error {
	return copyViaCommand("wl-copy", nil, text)
}

func copyViaXclip(text string) error {
	return copyViaCommand("xclip", []string{"-selection", "clipboard"}, text)
}

func copyViaXsel(text string) error {
	return copyViaCommand("xsel", []string{"--clipboard", "--input"}, text)
}

func copyViaAtotto(text string) error {
	if clipboard.Unsupported {
		return fmt.Errorf("atotto clipboard unsupported")
	}
	return clipboard.WriteAll(text)
}

func copyViaOSC52(text string) error {
	seq := osc52.New(text)
	switch {
	case os.Getenv("TMUX") != "":
		seq = seq.Tmux()
	case os.Getenv("STY") != "":
		seq = seq.Screen()
	}
	out := io.Writer(os.Stderr)
	if tty, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0); err == nil {
		out = tty
		defer tty.Close()
	}
	if _, err := fmt.Fprint(out, seq); err != nil {
		return fmt.Errorf("terminal clipboard: %w", err)
	}
	return nil
}

// copyText tries the system clipboard and always writes ~/.opentmd/last-copy.txt.
func copyText(text string) (clipboardOK bool, fallbackPath string, err error) {
	path, err := writeCopyFallback(text)
	if err != nil {
		return false, "", err
	}
	if err := copyToSystemClipboard(text); err == nil {
		return true, path, nil
	}
	return false, path, nil
}

func (m *Model) lastAssistantText() string {
	for i := len(m.messages) - 1; i >= 0; i-- {
		if m.messages[i].kind == "assistant" {
			if t := strings.TrimSpace(m.messages[i].text); t != "" {
				return t
			}
		}
	}
	return ""
}

func (m *Model) lastNonSystemText() string {
	for i := len(m.messages) - 1; i >= 0; i-- {
		if m.messages[i].kind == "system" {
			continue
		}
		if t := strings.TrimSpace(m.messages[i].text); t != "" {
			return t
		}
	}
	return ""
}

func (m *Model) conversationPlainText() string {
	var sb strings.Builder
	for _, msg := range m.messages {
		switch msg.kind {
		case "user":
			sb.WriteString("You: ")
			sb.WriteString(msg.text)
		case "assistant":
			sb.WriteString("Assistant: ")
			sb.WriteString(msg.text)
		case "tool":
			sb.WriteString("[tool] ")
			sb.WriteString(msg.text)
		default:
			sb.WriteString(msg.text)
		}
		sb.WriteByte('\n')
	}
	return strings.TrimSpace(sb.String())
}

func (m *Model) reportCopyResult(text string, clipOK bool, fallbackPath string, err error) {
	n := len([]rune(text))
	if err != nil {
		m.appendSystem("✗ 复制失败: " + err.Error())
		return
	}
	if clipOK {
		m.appendSystem(fmt.Sprintf("✓ 已复制到剪贴板（%d 字符）\n  备份: %s", n, fallbackPath))
		return
	}
	m.appendSystem(fmt.Sprintf(
		"剪贴板不可用（常见于 SSH 无图形环境），已写入文件（%d 字符）:\n  %s\n  可用 cat 查看，或在终端中用鼠标选中历史输出",
		n, fallbackPath,
	))
}

func (m *Model) handleCopyCommand(args []string) {
	switch {
	case len(args) == 0 || args[0] == "last":
		text := m.lastAssistantText()
		if text == "" {
			text = m.lastNonSystemText()
		}
		if text == "" {
			m.appendSystem("没有可复制的内容（等待 AI 回复，或使用 /copy all）")
			return
		}
		clipOK, path, err := copyText(text)
		m.reportCopyResult(text, clipOK, path, err)

	case args[0] == "all":
		text := m.conversationPlainText()
		if text == "" {
			m.appendSystem("当前会话为空")
			return
		}
		clipOK, path, err := copyText(text)
		m.reportCopyResult(text, clipOK, path, err)

	default:
		m.appendSystem("用法: /copy  |  /copy last  |  /copy all")
	}
}

func (m *Model) handleExportCommand(args []string) {
	text := m.conversationPlainText()
	if text == "" {
		m.appendSystem("当前会话为空")
		return
	}

	path := ""
	if len(args) > 0 {
		path = args[0]
	} else {
		name := fmt.Sprintf("opentmd-export-%s.txt", time.Now().Format("20060102-150405"))
		path = filepath.Join(m.agent.WorkDir(), name)
	}

	if err := os.WriteFile(path, []byte(text+"\n"), 0o644); err != nil {
		m.appendSystem("✗ 导出失败: " + err.Error())
		return
	}
	m.appendSystem(fmt.Sprintf("✓ 已导出到 %s（%d 字符）", path, len([]rune(text))))
}
