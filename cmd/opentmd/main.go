package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/opentmd/opentmd-cli/internal/agent"
	"github.com/opentmd/opentmd-cli/internal/cliname"
	"github.com/opentmd/opentmd-cli/internal/config"
	"github.com/opentmd/opentmd-cli/internal/project"
	"github.com/opentmd/opentmd-cli/internal/session"
	"github.com/opentmd/opentmd-cli/internal/setup"
	"github.com/opentmd/opentmd-cli/internal/tui"
	"github.com/opentmd/opentmd-cli/internal/trust"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	promptFlag     string
	promptFileFlag string
	dirFlag        string
	continueFlag   bool
	verboseFlag    bool
	modelFlag      string
	providerFlag   string
	loginProvider  string
	loginOAuth     bool
	trustFlag      bool
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   cliname.Current(),
	Short: "OpenTMD - AI coding assistant for the terminal",
	Long:  "opentmd-cli is an open-source AI coding agent that runs in your terminal.",
	RunE:  runDefault,
}

func init() {
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		return config.Ensure()
	}
	rootCmd.Flags().StringVarP(&promptFlag, "prompt", "p", "", "Run in prompt mode (non-interactive)")
	rootCmd.Flags().StringVar(&promptFileFlag, "prompt-file", "", "Read prompt from file")
	rootCmd.Flags().StringVarP(&dirFlag, "dir", "C", "", "Working directory")
	rootCmd.Flags().BoolVarP(&continueFlag, "continue", "c", false, "Continue the most recent session")
	rootCmd.Flags().BoolVarP(&verboseFlag, "verbose", "v", false, "Verbose output (tool calls to stderr)")
	rootCmd.Flags().StringVar(&modelFlag, "model", "", "Override model for this run")
	rootCmd.Flags().StringVar(&providerFlag, "provider", "", "Override provider for this run")
	rootCmd.Flags().BoolVar(&trustFlag, "trust", false, "Trust this workspace without prompting")
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(mcpCmd)
	rootCmd.AddCommand(daemonCmd)

	loginCmd.Flags().StringVar(&loginProvider, "provider", "", "Provider name")
	loginCmd.Flags().BoolVar(&loginOAuth, "oauth", false, "Login via OpenTMD OAuth")
}

func runDefault(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	applyOverrides(cfg)

	workDir, err := project.ResolveWorkDir(dirFlag)
	if err != nil {
		return err
	}
	if err := project.InitWorkspace(workDir, cfg); err != nil {
		return fmt.Errorf("init workspace: %w", err)
	}

	trustOpts := trustOptions()

	prompt := promptFlag
	if promptFileFlag != "" {
		data, err := os.ReadFile(promptFileFlag)
		if err != nil {
			return fmt.Errorf("read prompt file: %w", err)
		}
		prompt = string(data)
	}

	if prompt == "" && cfg.NeedsWizard() {
		if !term.IsTerminal(int(os.Stdin.Fd())) {
			return fmt.Errorf("provider not configured; run: %s login", cliname.Current())
		}
		if err := setup.RunWizard(cfg); err != nil {
			return err
		}
	}

	sess, err := session.NewStoreWithOptions(cfg, session.Options{ContinueLast: continueFlag})
	if err != nil {
		return fmt.Errorf("init session: %w", err)
	}

	opts := agent.Options{WorkDir: workDir, Verbose: verboseFlag}

	if prompt != "" {
		return runPromptMode(cfg, sess, prompt, opts, trustOpts)
	}

	ag, err := agent.New(cfg, sess, opts)
	if err != nil {
		return err
	}

	return tui.Run(ag, trustOpts)
}

func trustOptions() trust.Options {
	return trust.Options{
		Skip:      os.Getenv("OPENTMD_SKIP_TRUST") == "1",
		AutoTrust: trustFlag,
	}
}

func applyOverrides(cfg *config.Config) {
	if providerFlag != "" {
		if p, ok := config.PresetByName(providerFlag); ok {
			cfg.ApplyPreset(p)
		} else {
			cfg.Default.Provider = providerFlag
		}
	}
	if modelFlag != "" {
		cfg.Default.Model = modelFlag
	}
}

func runPromptMode(cfg *config.Config, sess *session.Store, prompt string, opts agent.Options, trustOpts trust.Options) error {
	if cfg.NeedsWizard() {
		return fmt.Errorf("provider not configured; run: %s login", cliname.Current())
	}
	if err := trust.EnsureTerminal(opts.WorkDir, trustOpts); err != nil {
		return err
	}

	ag, err := agent.New(cfg, sess, opts)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if opts.Verbose {
		printPromptBanner(cfg, opts.WorkDir, os.Stderr)
	}

	_, err = ag.ChatToWriter(ctx, prompt, os.Stdout)
	if err != nil && err != context.Canceled {
		return err
	}
	if isTTYStdout() {
		fmt.Println()
	}
	return nil
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Configure API key for a provider",
	RunE: func(cmd *cobra.Command, args []string) error {
		if loginOAuth {
			return setup.RunOAuth()
		}
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		return setup.RunLogin(cfg, loginProvider)
	},
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		path, _ := config.Path()
		if isTTYStdout() {
			printConfigHeader(path)
			printConfigRow("provider", cfg.Default.Provider)
			printConfigRow("model", cfg.Default.Model)
			printConfigRow("theme", cfg.TUI.Theme)
			printConfigRow("stream", fmt.Sprintf("%v", cfg.TUI.Stream))
			printConfigRow("persist", fmt.Sprintf("%v", cfg.Session.Persist))
			printConfigRow("workdir", cfg.DefaultWorkdir)
			printConfigRow("setup", fmt.Sprintf("completed=%v", cfg.Setup.Completed))
			printConfigRow("lsp", fmt.Sprintf("enabled=%v auto_detect=%v settle_ms=%d",
				cfg.LSP.Enabled, cfg.LSP.AutoDetect, cfg.LSP.DiagnosticsSettleDelayMs))
		} else {
			fmt.Printf("Config file: %s\n\n", path)
			fmt.Printf("  provider: %s\n", cfg.Default.Provider)
			fmt.Printf("  model:    %s\n", cfg.Default.Model)
			fmt.Printf("  theme:    %s\n", cfg.TUI.Theme)
			fmt.Printf("  stream:   %v\n", cfg.TUI.Stream)
			fmt.Printf("  persist:  %v\n", cfg.Session.Persist)
			fmt.Printf("  workdir:  %s\n", cfg.DefaultWorkdir)
			fmt.Printf("  setup:    completed=%v\n", cfg.Setup.Completed)
			fmt.Printf("  lsp:      enabled=%v auto_detect=%v settle_ms=%d\n",
				cfg.LSP.Enabled, cfg.LSP.AutoDetect, cfg.LSP.DiagnosticsSettleDelayMs)
		}

		name, pc, err := cfg.ActiveProvider()
		if err != nil {
			return err
		}
		keyStatus := "not set"
		if pc.APIKey != "" && pc.APIKey != "ollama" {
			keyStatus = "set (" + maskKey(pc.APIKey) + ")"
		} else if name == "ollama" {
			keyStatus = "not required (local)"
		}
		if isTTYStdout() {
			fmt.Println()
			fmt.Println(valueStyle().Bold(true).Render("[" + name + "] active"))
			printConfigRow("base_url", pc.BaseURL)
			printConfigRow("api_key", keyStatus)
			fmt.Println()
			fmt.Println(labelStyle().Render("Available providers:"))
		} else {
			fmt.Printf("\n  [%s] (active)\n", name)
			fmt.Printf("    base_url: %s\n", pc.BaseURL)
			fmt.Printf("    api_key:  %s\n", keyStatus)
			fmt.Println("\n  Available providers:")
		}

		for _, p := range config.ProviderPresets {
			pc := cfg.Providers[p.Name]
			status := "not configured"
			if p.Name == name {
				status = "active"
			} else if pc.APIKey != "" || !p.NeedAPIKey {
				status = "configured"
			}
			if isTTYStdout() {
				line := fmt.Sprintf("  %-10s %s [%s]", p.Name, p.Label, status)
				fmt.Println(valueStyle().Render(line))
			} else {
				fmt.Printf("    %-10s %s [%s]\n", p.Name, p.Label, status)
			}
		}
		return nil
	},
}

func maskKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "..." + key[len(key)-4:]
}
