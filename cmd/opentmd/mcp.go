package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/opentmd/opentmd/internal/config"
	"github.com/opentmd/opentmd/internal/mcp"
	"github.com/spf13/cobra"
)

var (
	mcpGlobalFlag bool
	mcpCmdFlag    string
	mcpArgsFlag   []string
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Manage MCP server configuration",
}

var mcpListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured MCP servers",
	RunE: func(cmd *cobra.Command, args []string) error {
		workDir, _ := os.Getwd()
		cfg, err := mcp.LoadConfig(workDir)
		if err != nil {
			return err
		}
		if len(cfg.MCPServers) == 0 {
			fmt.Println("No MCP servers configured.")
			path, _ := mcp.DefaultConfigPath(true, workDir)
			fmt.Printf("Edit %s to add servers.\n", path)
			return nil
		}
		for name, sc := range cfg.MCPServers {
			status := "enabled"
			if sc.Disabled {
				status = "disabled"
			}
			fmt.Printf("  %s [%s] command=%s %s\n", name, status, sc.Command, strings.Join(sc.Args, " "))
		}
		return nil
	},
}

var mcpAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add or update an MCP server",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if mcpCmdFlag == "" {
			return fmt.Errorf("--command is required")
		}
		workDir, _ := os.Getwd()
		path, err := mcp.DefaultConfigPath(mcpGlobalFlag, workDir)
		if err != nil {
			return err
		}
		sc := mcp.ServerConfig{
			Command: mcpCmdFlag,
			Args:    mcpArgsFlag,
		}
		if err := mcp.SaveServer(path, args[0], sc); err != nil {
			return err
		}
		fmt.Printf("Saved MCP server %q to %s\n", args[0], path)
		return nil
	},
}

var mcpShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show raw mcp.json",
	RunE: func(cmd *cobra.Command, args []string) error {
		workDir, _ := os.Getwd()
		path, err := mcp.DefaultConfigPath(mcpGlobalFlag, workDir)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			dir, _ := config.Dir()
			path = dir + "/mcp.json"
			data, err = os.ReadFile(path)
		}
		if err != nil {
			return fmt.Errorf("read mcp config: %w", err)
		}
		var pretty json.RawMessage
		if json.Unmarshal(data, &pretty) == nil {
			out, _ := json.MarshalIndent(pretty, "", "  ")
			fmt.Println(string(out))
			return nil
		}
		fmt.Println(string(data))
		return nil
	},
}

func init() {
	mcpAddCmd.Flags().BoolVar(&mcpGlobalFlag, "global", true, "Write to ~/.opentmd/mcp.json")
	mcpAddCmd.Flags().StringVar(&mcpCmdFlag, "command", "", "MCP server command")
	mcpAddCmd.Flags().StringSliceVar(&mcpArgsFlag, "args", nil, "MCP server arguments")

	mcpShowCmd.Flags().BoolVar(&mcpGlobalFlag, "global", true, "Show global mcp.json")

	mcpCmd.AddCommand(mcpListCmd, mcpAddCmd, mcpShowCmd)
}
