package setup

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/opentmd/opentmd/internal/config"
	"golang.org/x/term"
)

// RunWizard runs the 3-step first-run setup wizard.
func RunWizard(cfg *config.Config) error {
	in := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Println("  ┌──────────────────────────────────────┐")
	fmt.Println("  │       OpenTMD 首次启动向导 (1/3)      │")
	fmt.Println("  └──────────────────────────────────────┘")
	fmt.Println()
	fmt.Println("  欢迎使用 OpenTMD — 终端 AI 编码助手")
	fmt.Println("  请选择大模型 Provider：")
	fmt.Println()

	for i, p := range config.ProviderPresets {
		fmt.Printf("    %d) %s (%s)\n", i+1, p.Label, p.Name)
	}
	fmt.Println("    0) 退出")
	fmt.Println()

	choice, err := promptLine(in, "  请输入序号 [1]: ")
	if err != nil {
		return err
	}
	if choice == "0" {
		return fmt.Errorf("setup cancelled")
	}
	if choice == "" {
		choice = "1"
	}
	idx, err := strconv.Atoi(choice)
	if err != nil || idx < 1 || idx > len(config.ProviderPresets) {
		return fmt.Errorf("invalid choice: %s", choice)
	}
	preset := config.ProviderPresets[idx-1]
	cfg.ApplyPreset(preset)

	fmt.Println()
	fmt.Printf("  ┌──────────────────────────────────────┐\n")
	fmt.Printf("  │       OpenTMD 首次启动向导 (2/3)      │\n")
	fmt.Printf("  └──────────────────────────────────────┘\n")
	fmt.Println()
	fmt.Printf("  Provider: %s\n", preset.Label)
	fmt.Printf("  API 地址: %s\n", preset.BaseURL)
	fmt.Println()

	if preset.NeedAPIKey {
		fmt.Println("  请输入 API Key（输入不可见）：")
		key, err := readSecret()
		if err != nil {
			return err
		}
		key = strings.TrimSpace(key)
		if key == "" {
			return fmt.Errorf("API key cannot be empty")
		}
		if err := cfg.SetAPIKey(preset.Name, key); err != nil {
			return err
		}
	} else {
		fmt.Println("  Ollama 本地模型无需 API Key，请确保 ollama serve 已运行。")
		if err := cfg.SetAPIKey(preset.Name, "ollama"); err != nil {
			return err
		}
	}

	fmt.Println()
	fmt.Printf("  ┌──────────────────────────────────────┐\n")
	fmt.Printf("  │       OpenTMD 首次启动向导 (3/3)      │\n")
	fmt.Printf("  └──────────────────────────────────────┘\n")
	fmt.Println()
	fmt.Printf("  默认模型 [%s]: ", preset.DefaultModel)
	model, err := promptLine(in, "")
	if err != nil {
		return err
	}
	if model == "" {
		model = preset.DefaultModel
	}
	cfg.Default.Model = model

	if err := config.Save(cfg); err != nil {
		return err
	}
	if err := cfg.MarkSetupCompleted(); err != nil {
		return err
	}

	path, _ := config.Path()
	fmt.Println()
	fmt.Println("  ✓ 配置完成！")
	fmt.Printf("  配置文件: %s\n", path)
	fmt.Printf("  Provider: %s\n", cfg.Default.Provider)
	fmt.Printf("  Model:    %s\n", cfg.Default.Model)
	fmt.Println()
	fmt.Println("  进入项目目录后运行: opentmd")
	fmt.Println()

	return nil
}

// RunLogin configures a provider interactively (used by `opentmd login`).
func RunLogin(cfg *config.Config, providerName string) error {
	in := bufio.NewReader(os.Stdin)

	var preset config.ProviderPreset
	var ok bool

	if providerName != "" {
		preset, ok = config.PresetByName(providerName)
		if !ok {
			return fmt.Errorf("unknown provider %q", providerName)
		}
	} else {
		fmt.Println("选择 Provider：")
		for i, p := range config.ProviderPresets {
			fmt.Printf("  %d) %s\n", i+1, p.Label)
		}
		choice, err := promptLine(in, "序号 [1]: ")
		if err != nil {
			return err
		}
		if choice == "" {
			choice = "1"
		}
		idx, err := strconv.Atoi(choice)
		if err != nil || idx < 1 || idx > len(config.ProviderPresets) {
			return fmt.Errorf("invalid choice")
		}
		preset = config.ProviderPresets[idx-1]
	}

	cfg.ApplyPreset(preset)

	if preset.NeedAPIKey {
		fmt.Printf("Provider: %s\n", preset.Label)
		fmt.Print("Enter API key: ")
		key, err := readSecret()
		if err != nil {
			return err
		}
		key = strings.TrimSpace(key)
		if key == "" {
			return fmt.Errorf("API key cannot be empty")
		}
		if err := cfg.SetAPIKey(preset.Name, key); err != nil {
			return err
		}
	} else {
		if err := cfg.SetAPIKey(preset.Name, "ollama"); err != nil {
			return err
		}
	}

	if err := cfg.MarkSetupCompleted(); err != nil {
		return err
	}
	fmt.Println("API key saved successfully.")
	return nil
}

// RunOAuth shows OpenTMD OAuth placeholder (future feature).
func RunOAuth() error {
	fmt.Println("OpenTMD OAuth 登录即将上线。")
	fmt.Println("当前请使用: opentmd login")
	return fmt.Errorf("oauth not available yet")
}

func promptLine(in *bufio.Reader, prompt string) (string, error) {
	if prompt != "" {
		fmt.Print(prompt)
	}
	line, err := in.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func readSecret() (string, error) {
	fd := int(os.Stdin.Fd())
	if term.IsTerminal(fd) {
		b, err := term.ReadPassword(fd)
		if err != nil {
			return "", err
		}
		fmt.Println()
		return string(b), nil
	}
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return scanner.Text(), nil
	}
	return "", scanner.Err()
}
