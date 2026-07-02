package tool

import (
	"context"
	"encoding/json"

	"github.com/opentmd/opentmd/internal/llm"
	"github.com/opentmd/opentmd/internal/semantic"
	"github.com/opentmd/opentmd/internal/web"
)

func (r *Runtime) registerWebTools() {
	r.Register(llm.ToolDefinition{
		Name: "web_search",
		Description: "Search the web for documentation, APIs, libraries, or information not available locally.",
		Parameters: objParam(map[string]any{
			"query":       strParam("Search query"),
			"max_results": intParam("Max results (default 8, max 20)"),
		}, "query"),
	}, r.webSearch)

	r.Register(llm.ToolDefinition{
		Name: "web_fetch",
		Description: "Fetch a public http(s) URL and return readable text content.",
		Parameters: objParam(map[string]any{
			"url":       strParam("URL to fetch"),
			"max_chars": intParam("Max characters to return (default 20000)"),
		}, "url"),
	}, r.webFetch)
}

func (r *Runtime) registerSemanticTools() {
	r.Register(llm.ToolDefinition{
		Name: "list_symbols",
		Description: "List top-level symbols (functions, classes, types) in a source file with line ranges.",
		Parameters: objParam(map[string]any{
			"path": strParam("Source file path"),
		}, "path"),
	}, r.listSymbols)

	r.Register(llm.ToolDefinition{
		Name: "read_symbol",
		Description: "Read a named symbol from a source file with surrounding context lines.",
		Parameters: objParam(map[string]any{
			"path":          strParam("Source file path"),
			"symbol":        strParam("Symbol name"),
			"context_lines": intParam("Context lines around symbol (default 8)"),
		}, "path", "symbol"),
	}, r.readSymbol)
}

func (r *Runtime) registerDiagnosticsTool() {
	r.Register(llm.ToolDefinition{
		Name: "diagnostics",
		Description: "Get real-time compiler/linter diagnostics from Language Server Protocol (gopls, rust-analyzer, typescript-language-server, pylsp, etc.).",
		Parameters: objParam(map[string]any{
			"path":     strParam("Optional file path to check"),
			"severity": strParam("Filter: error, warning, or all (default error)"),
		}),
	}, r.diagnosticsTool)
}

type webSearchArgs struct {
	Query      string `json:"query"`
	MaxResults int    `json:"max_results"`
}

func (r *Runtime) webSearch(ctx context.Context, args json.RawMessage) (string, error) {
	a, err := parseArgs[webSearchArgs](args, "web_search")
	if err != nil {
		return "", err
	}
	return web.Search(ctx, a.Query, a.MaxResults)
}

type webFetchArgs struct {
	URL      string `json:"url"`
	MaxChars int    `json:"max_chars"`
}

func (r *Runtime) webFetch(ctx context.Context, args json.RawMessage) (string, error) {
	a, err := parseArgs[webFetchArgs](args, "web_fetch")
	if err != nil {
		return "", err
	}
	return web.Fetch(ctx, a.URL, a.MaxChars)
}

type pathOnlyArgs struct {
	Path string `json:"path"`
}

func (r *Runtime) listSymbols(ctx context.Context, args json.RawMessage) (string, error) {
	a, err := parseArgs[pathOnlyArgs](args, "list_symbols")
	if err != nil {
		return "", err
	}
	path, err := r.resolve(a.Path)
	if err != nil {
		return "", err
	}
	symbols, err := semantic.ListSymbols(path)
	if err != nil {
		return "", err
	}
	return semantic.FormatSymbols(a.Path, symbols), nil
}

type readSymbolArgs struct {
	Path         string `json:"path"`
	Symbol       string `json:"symbol"`
	ContextLines int    `json:"context_lines"`
}

func (r *Runtime) readSymbol(ctx context.Context, args json.RawMessage) (string, error) {
	a, err := parseArgs[readSymbolArgs](args, "read_symbol")
	if err != nil {
		return "", err
	}
	path, err := r.resolve(a.Path)
	if err != nil {
		return "", err
	}
	return semantic.ReadSymbol(path, a.Symbol, a.ContextLines)
}

type diagnosticsArgs struct {
	Path     string `json:"path"`
	Severity string `json:"severity"`
}

func (r *Runtime) diagnosticsTool(ctx context.Context, args json.RawMessage) (string, error) {
	a, err := parseArgs[diagnosticsArgs](args, "diagnostics")
	if err != nil {
		return "", err
	}
	if r.lsp == nil {
		return "LSP not available.", nil
	}
	return r.lsp.Collect(ctx, a.Path, a.Severity)
}
