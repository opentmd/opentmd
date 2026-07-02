package main

import (
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/opentmd/opentmd/internal/config"
)

func printPromptBanner(cfg *config.Config, workDir string, w io.Writer) {
	if !lipgloss.HasDarkBackground() && cfg.TUI.Theme != "light" {
		// still print plain banner on light terminals
		fmt.Fprintf(w, "→ OpenTMD  %s/%s  %s\n\n", cfg.Default.Provider, cfg.Default.Model, workDir)
		return
	}
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#A78BFA")).Render("OpenTMD")
	meta := lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4")).Render(
		fmt.Sprintf("%s/%s · %s", cfg.Default.Provider, cfg.Default.Model, workDir),
	)
	fmt.Fprintf(w, "%s  %s\n\n", title, meta)
}

func printConfigHeader(path string) {
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#A78BFA")).Render("OpenTMD 配置")
	pathStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4"))
	fmt.Println(title)
	fmt.Println(pathStyle.Render("文件: " + path))
	fmt.Println()
}

func labelStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4")).Width(12)
}

func valueStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#F8F8F2"))
}

func printConfigRow(label, value string) {
	if !isTTYStdout() {
		fmt.Printf("  %-12s %s\n", label+":", value)
		return
	}
	fmt.Println(labelStyle().Render(label) + valueStyle().Render(value))
}

func isTTYStdout() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}
