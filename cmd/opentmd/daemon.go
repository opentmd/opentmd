package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/opentmd/opentmd/internal/cliname"
	"github.com/opentmd/opentmd/internal/config"
	"github.com/opentmd/opentmd/internal/daemon"
	"github.com/opentmd/opentmd/internal/project"
	"github.com/opentmd/opentmd/internal/trust"
	"github.com/spf13/cobra"
)

var daemonPort int
var daemonDir string

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Run OpenTMD HTTP daemon for IDE integration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if cfg.NeedsWizard() {
			return fmt.Errorf("provider not configured; run: %s login", cliname.Current())
		}
		workDir, err := project.ResolveWorkDir(daemonDir)
		if err != nil {
			return err
		}
		if err := project.InitWorkspace(workDir, cfg); err != nil {
			return fmt.Errorf("init workspace: %w", err)
		}
		if err := trust.EnsureTerminal(workDir, trustOptions()); err != nil {
			return err
		}
		srv := daemon.New(cfg, daemonPort)
		srv.SetWorkDir(workDir)

		errCh := make(chan error, 1)
		go func() {
			errCh <- srv.ListenAndServe()
		}()

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		select {
		case sig := <-sigCh:
			fmt.Fprintf(os.Stderr, "\nreceived %v, shutting down…\n", sig)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			return srv.Shutdown(ctx)
		case err := <-errCh:
			return err
		}
	},
}

func init() {
	daemonCmd.Flags().IntVarP(&daemonPort, "port", "p", daemon.DefaultPort, "Listen port")
	daemonCmd.Flags().StringVarP(&daemonDir, "dir", "C", "", "Default working directory")
}
