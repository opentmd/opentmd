package tui

import "github.com/charmbracelet/lipgloss"

// Palette uses 16-color SGR strategy for terminal theme adaptability.
type Theme struct {
	Name       string
	Primary    lipgloss.Color // brand — magenta
	Secondary  lipgloss.Color // body text
	Accent     lipgloss.Color // cyan — chevrons, borders, code
	User       lipgloss.Color
	Assistant  lipgloss.Color
	System     lipgloss.Color
	Tool       lipgloss.Color
	Warn       lipgloss.Color
	Muted      lipgloss.Color
	Border     lipgloss.Color
	InputBorder lipgloss.Color
	Success    lipgloss.Color
	ReverseFG  lipgloss.Color
	ReverseBG  lipgloss.Color
}

const appVersion = "0.1.3"

func loadTheme(name string) Theme {
	if name == "light" {
		return Theme{
			Name:        "light",
			Primary:     lipgloss.Color("5"),
			Secondary:   lipgloss.Color("0"),
			Accent:      lipgloss.Color("6"),
			User:        lipgloss.Color("6"),
			Assistant:   lipgloss.Color("0"),
			System:      lipgloss.Color("8"),
			Tool:        lipgloss.Color("3"),
			Warn:        lipgloss.Color("9"),
			Muted:       lipgloss.Color("8"),
			Border:      lipgloss.Color("6"),
			InputBorder: lipgloss.Color("6"),
			Success:     lipgloss.Color("2"),
			ReverseFG:   lipgloss.Color("0"),
			ReverseBG:   lipgloss.Color("6"),
		}
	}
	return Theme{
		Name:        "dark",
		Primary:     lipgloss.Color("5"),
		Secondary:   lipgloss.Color("7"),
		Accent:      lipgloss.Color("6"),
		User:        lipgloss.Color("6"),
		Assistant:   lipgloss.Color("7"),
		System:      lipgloss.Color("8"),
		Tool:        lipgloss.Color("3"),
		Warn:        lipgloss.Color("9"),
		Muted:       lipgloss.Color("7"),
		Border:      lipgloss.Color("6"),
		InputBorder: lipgloss.Color("6"),
		Success:     lipgloss.Color("10"),
		ReverseFG:   lipgloss.Color("0"),
		ReverseBG:   lipgloss.Color("6"),
	}
}

func (t Theme) inputStyles() textareaStyle {
	return textareaStyle{
		focused: textareaStyleConfig{
			prompt:      lipgloss.NewStyle().Foreground(t.Accent).Bold(true),
			text:        lipgloss.NewStyle().Foreground(t.Secondary),
			placeholder: lipgloss.NewStyle().Foreground(t.Muted).Faint(true),
			cursorLine:  lipgloss.NewStyle(),
		},
		blurred: textareaStyleConfig{
			prompt:      lipgloss.NewStyle().Foreground(t.Muted),
			text:        lipgloss.NewStyle().Foreground(t.Muted),
			placeholder: lipgloss.NewStyle().Foreground(t.Muted).Faint(true),
		},
	}
}

type textareaStyle struct {
	focused, blurred textareaStyleConfig
}

type textareaStyleConfig struct {
	prompt, text, placeholder, cursorLine lipgloss.Style
}

func (t Theme) styles() themeStyles {
	s := themeStyles{
		title: lipgloss.NewStyle().Bold(true).Foreground(t.Primary),
		subtitle: lipgloss.NewStyle().Foreground(t.Muted).Faint(true),
		user: lipgloss.NewStyle().Foreground(t.Secondary),
		userChevron: lipgloss.NewStyle().Foreground(t.Accent).Bold(true),
		assistant: lipgloss.NewStyle().Foreground(t.Assistant),
		system: lipgloss.NewStyle().Foreground(t.System).Faint(true),
		tool: lipgloss.NewStyle().Foreground(t.Secondary),
		toolBullet: lipgloss.NewStyle().Foreground(t.Accent).Bold(true),
		warn: lipgloss.NewStyle().Foreground(t.Warn).Bold(true),
		footer: lipgloss.NewStyle().Foreground(t.Muted).Faint(true),
		input: lipgloss.NewStyle().Foreground(t.Secondary),
		viewport: lipgloss.NewStyle(),
		welcome: lipgloss.NewStyle().Foreground(t.Muted),
		hint: lipgloss.NewStyle().Foreground(t.Accent),
		busy: lipgloss.NewStyle().Foreground(t.Accent).Bold(true),
		success: lipgloss.NewStyle().Foreground(t.Success),
		error:   lipgloss.NewStyle().Foreground(t.Warn).Bold(true),
		code: lipgloss.NewStyle().Foreground(t.Accent),
		codeInline: lipgloss.NewStyle().Foreground(t.Accent),
		toolHeader: lipgloss.NewStyle().Foreground(t.Muted).Faint(true),
		inputBar: lipgloss.NewStyle().Foreground(t.Muted).Faint(true),
		rule: lipgloss.NewStyle().Foreground(t.Border),
		menuActive: lipgloss.NewStyle().
			Background(t.ReverseBG).
			Foreground(t.ReverseFG).
			Bold(true),
		menuInactive: lipgloss.NewStyle().Foreground(t.Secondary),
		menuDesc: lipgloss.NewStyle().Foreground(t.Muted).Faint(true),
		bullet: lipgloss.NewStyle().Foreground(t.Accent),
	}
	return s
}

type themeStyles struct {
	title, subtitle, user, userChevron, assistant, system, tool, toolBullet, warn, footer, input, viewport, welcome, hint, busy lipgloss.Style
	success, error, code, codeInline, toolHeader, inputBar, rule lipgloss.Style
	menuActive, menuInactive, menuDesc, bullet lipgloss.Style
}
