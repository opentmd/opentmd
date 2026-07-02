package semantic

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type Symbol struct {
	Name      string
	Kind      string
	StartLine int
	EndLine   int
}

var regexPatterns = map[string][]struct {
	kind string
	re   *regexp.Regexp
}{
	"go": {
		{"func", regexp.MustCompile(`^func\s+(?:\([^)]+\)\s+)?([A-Za-z_]\w*)\s*\(`)},
		{"type", regexp.MustCompile(`^type\s+([A-Za-z_]\w*)\s`)},
		{"const", regexp.MustCompile(`^const\s+([A-Za-z_]\w*)\s`)},
		{"var", regexp.MustCompile(`^var\s+([A-Za-z_]\w*)\s`)},
	},
	"rust": {
		{"fn", regexp.MustCompile(`^pub\s+fn\s+([A-Za-z_]\w*)`)},
		{"fn", regexp.MustCompile(`^fn\s+([A-Za-z_]\w*)`)},
		{"struct", regexp.MustCompile(`^pub\s+struct\s+([A-Za-z_]\w*)`)},
		{"struct", regexp.MustCompile(`^struct\s+([A-Za-z_]\w*)`)},
		{"enum", regexp.MustCompile(`^pub\s+enum\s+([A-Za-z_]\w*)`)},
		{"impl", regexp.MustCompile(`^impl(?:<[^>]+>)?\s+([A-Za-z_]\w*)`)},
	},
	"python": {
		{"class", regexp.MustCompile(`^class\s+([A-Za-z_]\w*)`)},
		{"def", regexp.MustCompile(`^def\s+([A-Za-z_]\w*)\s*\(`)},
		{"async def", regexp.MustCompile(`^async\s+def\s+([A-Za-z_]\w*)\s*\(`)},
	},
	"javascript": jsPatterns(),
	"typescript": jsPatterns(),
	"java": {
		{"class", regexp.MustCompile(`^(?:public\s+|private\s+|protected\s+)?(?:static\s+)?class\s+([A-Za-z_]\w*)`)},
		{"interface", regexp.MustCompile(`^(?:public\s+)?interface\s+([A-Za-z_]\w*)`)},
		{"method", regexp.MustCompile(`^(?:public|private|protected)\s+(?:static\s+)?[\w<>,\[\]\s]+\s+([A-Za-z_]\w*)\s*\(`)},
	},
	"c": {
		{"function", regexp.MustCompile(`^[\w\s\*]+\s+([A-Za-z_]\w*)\s*\([^;]*\)\s*\{`)},
		{"struct", regexp.MustCompile(`^struct\s+([A-Za-z_]\w*)`)},
	},
	"cpp": {
		{"class", regexp.MustCompile(`^class\s+([A-Za-z_]\w*)`)},
		{"struct", regexp.MustCompile(`^struct\s+([A-Za-z_]\w*)`)},
		{"function", regexp.MustCompile(`^[\w:<>,\*\&\s]+\s+([A-Za-z_]\w*)\s*\([^;]*\)\s*\{`)},
	},
}

func jsPatterns() []struct {
	kind string
	re   *regexp.Regexp
} {
	return []struct {
		kind string
		re   *regexp.Regexp
	}{
		{"function", regexp.MustCompile(`^function\s+([A-Za-z_]\w*)\s*\(`)},
		{"class", regexp.MustCompile(`^class\s+([A-Za-z_]\w*)`)},
		{"const", regexp.MustCompile(`^(?:export\s+)?const\s+([A-Za-z_]\w*)\s*=`)},
		{"method", regexp.MustCompile(`^(?:async\s+)?([A-Za-z_]\w*)\s*\([^)]*\)\s*\{`)},
	}
}

func listSymbolsRegex(path string, source []byte, lang string) ([]Symbol, error) {
	pats, ok := regexPatterns[lang]
	if !ok {
		return nil, fmt.Errorf("no symbol patterns for %s", lang)
	}

	var symbols []Symbol
	sc := bufio.NewScanner(strings.NewReader(string(source)))
	lineNo := 0
	var current *Symbol
	for sc.Scan() {
		lineNo++
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "#") {
			continue
		}
		for _, p := range pats {
			m := p.re.FindStringSubmatch(line)
			if len(m) < 2 {
				continue
			}
			if current != nil {
				current.EndLine = lineNo - 1
				symbols = append(symbols, *current)
			}
			current = &Symbol{Name: m[1], Kind: p.kind, StartLine: lineNo, EndLine: lineNo}
			break
		}
	}
	if current != nil {
		current.EndLine = lineNo
		symbols = append(symbols, *current)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return symbols, nil
}

func ListSymbols(path string) ([]Symbol, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if syms, err := listSymbolsTreeSitter(path, data); err == nil && len(syms) > 0 {
		return syms, nil
	}

	lang := DetectLanguage(path)
	if lang == "" {
		return nil, fmt.Errorf("unsupported language for %s", path)
	}
	return listSymbolsRegex(path, data, lang)
}

func FormatSymbols(path string, symbols []Symbol) string {
	if len(symbols) == 0 {
		return "no symbols found in " + path
	}
	var sb strings.Builder
	sb.WriteString("Symbols in ")
	sb.WriteString(path)
	sb.WriteString(":\n")
	for _, s := range symbols {
		fmt.Fprintf(&sb, "  %s %s @ L%d-L%d\n", s.Kind, s.Name, s.StartLine, s.EndLine)
	}
	return strings.TrimSpace(sb.String())
}

func ReadSymbol(path, name string, contextLines int) (string, error) {
	if contextLines <= 0 {
		contextLines = 8
	}
	symbols, err := ListSymbols(path)
	if err != nil {
		return "", err
	}
	var target *Symbol
	for i := range symbols {
		if symbols[i].Name == name {
			target = &symbols[i]
			break
		}
	}
	if target == nil {
		return "", fmt.Errorf("symbol %q not found in %s", name, path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(data), "\n")
	start := target.StartLine - 1 - contextLines
	if start < 0 {
		start = 0
	}
	end := target.EndLine + contextLines
	if end > len(lines) {
		end = len(lines)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "%s %s in %s (L%d-L%d, showing L%d-L%d):\n\n",
		target.Kind, target.Name, path, target.StartLine, target.EndLine, start+1, end)
	for i := start; i < end; i++ {
		fmt.Fprintf(&sb, "%4d| %s\n", i+1, lines[i])
	}
	return sb.String(), nil
}
