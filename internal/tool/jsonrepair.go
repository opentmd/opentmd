package tool

import (
	"encoding/json"
	"regexp"
	"strings"
)

var (
	fenceRe       = regexp.MustCompile("(?s)^```(?:json)?\\s*")
	trailingComma = regexp.MustCompile(`,(\s*[}\]])`)
)

// RepairToolArgs normalizes malformed LLM tool-call JSON before execution.
func RepairToolArgs(toolName, args string) string {
	args = strings.TrimSpace(args)
	if args == "" {
		return args
	}
	pre := preEscapeWindowsPaths(args)
	if json.Valid([]byte(pre)) {
		return pre
	}
	repaired := repairJSON(pre)
	if json.Valid([]byte(repaired)) {
		return repaired
	}
	if toolName == "edit_file" {
		if v := extractEditFileArgs(pre); v != nil {
			if b, err := json.Marshal(v); err == nil {
				return string(b)
			}
		}
	}
	return args
}

func repairJSON(s string) string {
	s = strings.TrimSpace(s)
	s = fenceRe.ReplaceAllString(s, "")
	s = strings.TrimSuffix(strings.TrimSpace(s), "```")
	s = trailingComma.ReplaceAllString(s, "$1")
	s = strings.ReplaceAll(s, `'`, `"`)
	return strings.TrimSpace(s)
}

func preEscapeWindowsPaths(s string) string {
	// Fix D:\path style strings that JSON interprets as escape sequences.
	re := regexp.MustCompile(`"([A-Za-z]:\\[^"]*)"`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		inner := match[1 : len(match)-1]
		var b strings.Builder
		b.WriteByte('"')
		for i := 0; i < len(inner); i++ {
			if inner[i] == '\\' && i+1 < len(inner) {
				next := inner[i+1]
				if next == 't' || next == 'n' || next == 'r' || next == 'b' || next == 'f' || next == 'u' {
					b.WriteString(`\\`)
					continue
				}
			}
			b.WriteByte(inner[i])
		}
		b.WriteByte('"')
		return b.String()
	})
}

func extractEditFileArgs(s string) map[string]string {
	out := map[string]string{}
	for _, key := range []string{"path", "old_string", "new_string"} {
		re := regexp.MustCompile(`"` + key + `"\s*:\s*"((?:\\.|[^"\\])*)"`)
		if m := re.FindStringSubmatch(s); len(m) > 1 {
			out[key] = strings.ReplaceAll(m[1], `\"`, `"`)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
