package semantic

import (
	"embed"
	"fmt"
)

//go:embed queries/*.scm
var queryFS embed.FS

func symbolsQuery(lang string) (string, error) {
	file := queryLanguageFile(lang)
	data, err := queryFS.ReadFile("queries/" + file)
	if err != nil {
		return "", fmt.Errorf("no symbols query for %q", lang)
	}
	return string(data), nil
}

func queryLanguageFile(lang string) string {
	switch lang {
	case "tsx", "jsx":
		return "javascript.scm"
	case "c_sharp":
		return "csharp.scm"
	default:
		return lang + ".scm"
	}
}
