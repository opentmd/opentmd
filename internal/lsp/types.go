package lsp

import (
	"fmt"
	"sort"
	"strings"
)

type Severity int

const (
	SeverityError Severity = iota + 1
	SeverityWarning
	SeverityInfo
	SeverityHint
)

func severityFromLSP(v int) Severity {
	switch v {
	case 1:
		return SeverityError
	case 2:
		return SeverityWarning
	case 3:
		return SeverityInfo
	default:
		return SeverityHint
	}
}

func (s Severity) Label() string {
	switch s {
	case SeverityError:
		return "ERROR"
	case SeverityWarning:
		return "WARN"
	case SeverityInfo:
		return "INFO"
	default:
		return "HINT"
	}
}

type Diagnostic struct {
	File      string
	Line      int
	Column    int
	EndLine   int
	EndColumn int
	Severity  Severity
	Message   string
	Source    string
	Code      string
}

func (d Diagnostic) DisplayLine() string {
	code := ""
	if d.Code != "" {
		code = " " + d.Code
	}
	src := ""
	if d.Source != "" {
		src = fmt.Sprintf(" (%s)", d.Source)
	}
	return fmt.Sprintf("%s:%d:%d [%s]%s %s%s", d.File, d.Line, d.Column, d.Severity.Label(), code, d.Message, src)
}

func FormatDiagnostics(diags []Diagnostic, filter string) string {
	filtered := FilterDiagnostics(diags, filter)
	if len(filtered) == 0 {
		return ""
	}
	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].File != filtered[j].File {
			return filtered[i].File < filtered[j].File
		}
		if filtered[i].Line != filtered[j].Line {
			return filtered[i].Line < filtered[j].Line
		}
		return filtered[i].Column < filtered[j].Column
	})
	lines := make([]string, len(filtered))
	for i, d := range filtered {
		lines[i] = d.DisplayLine()
	}
	return strings.Join(lines, "\n")
}

func FilterDiagnostics(diags []Diagnostic, severity string) []Diagnostic {
	if severity == "" || severity == "all" {
		return diags
	}
	var out []Diagnostic
	for _, d := range diags {
		switch severity {
		case "warning":
			if d.Severity == SeverityWarning {
				out = append(out, d)
			}
		default:
			if d.Severity == SeverityError {
				out = append(out, d)
			}
		}
	}
	return out
}
