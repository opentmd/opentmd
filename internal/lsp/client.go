package lsp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

const requestTimeout = 30 * time.Second

type openDoc struct {
	uri        string
	languageID string
	version    int32
}

type Client struct {
	cmd    *exec.Cmd
	writer io.WriteCloser
	reader *bufio.Reader
	writeMu sync.Mutex

	nextID   atomic.Int64
	pending  sync.Map // id int64 -> chan json.RawMessage
	diags    sync.Map // path string -> []Diagnostic
	openDocs sync.Map // path string -> openDoc

	closeOnce sync.Once
	done      chan struct{}
}

func StartClient(ctx context.Context, cfg ServerConfig, projectRoot, languageID string) (*Client, error) {
	if !commandAvailable(cfg.Command) {
		return nil, fmt.Errorf("LSP command not found: %s", cfg.Command)
	}

	cmd := exec.CommandContext(ctx, cfg.Command, cfg.Args...)
	cmd.Dir = projectRoot
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	cmd.Stderr = io.Discard
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start LSP %s: %w", cfg.Command, err)
	}

	c := &Client{
		cmd:    cmd,
		writer: stdin,
		reader: bufio.NewReader(stdout),
		done:   make(chan struct{}),
	}
	go c.readLoop()

	rootURI := pathToURI(projectRoot)
	initParams := map[string]any{
		"processId": os.Getpid(),
		"rootUri":   rootURI,
		"capabilities": map[string]any{
			"textDocument": map[string]any{
				"publishDiagnostics": map[string]any{"relatedInformation": true},
				"synchronization":    map[string]any{"didOpen": true, "didChange": true},
			},
		},
		"clientInfo": map[string]any{"name": "opentmd", "version": "1.0"},
	}
	if _, err := c.request("initialize", initParams); err != nil {
		c.Close()
		return nil, fmt.Errorf("LSP initialize: %w", err)
	}
	if err := c.notify("initialized", map[string]any{}); err != nil {
		c.Close()
		return nil, err
	}
	_ = languageID
	return c, nil
}

func (c *Client) readLoop() {
	defer close(c.done)
	for {
		msg, err := ReadMessage(c.reader)
		if err != nil {
			return
		}
		c.dispatch(msg)
	}
}

func (c *Client) dispatch(raw json.RawMessage) {
	var envelope struct {
		ID     *json.RawMessage `json:"id"`
		Method string           `json:"method"`
		Result json.RawMessage  `json:"result"`
		Error  json.RawMessage  `json:"error"`
		Params json.RawMessage  `json:"params"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return
	}

	if envelope.ID != nil {
		id, ok := parseID(*envelope.ID)
		if !ok {
			return
		}
		if ch, ok := c.pending.Load(id); ok {
			c.pending.Delete(id)
			ch.(chan json.RawMessage) <- raw
		}
		return
	}

	if envelope.Method == "textDocument/publishDiagnostics" {
		c.handleDiagnostics(envelope.Params)
	}
}

func parseID(raw json.RawMessage) (int64, bool) {
	var f float64
	if err := json.Unmarshal(raw, &f); err == nil {
		return int64(f), true
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		var n int64
		_, err := fmt.Sscan(s, &n)
		return n, err == nil
	}
	return 0, false
}

func (c *Client) handleDiagnostics(params json.RawMessage) {
	var p struct {
		URI          string `json:"uri"`
		Diagnostics  []struct {
			Range struct {
				Start struct {
					Line      int `json:"line"`
					Character int `json:"character"`
				} `json:"start"`
				End struct {
					Line      int `json:"line"`
					Character int `json:"character"`
				} `json:"end"`
			} `json:"range"`
			Severity int    `json:"severity"`
			Message  string `json:"message"`
			Source   string `json:"source"`
			Code     any    `json:"code"`
		} `json:"diagnostics"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return
	}

	path := uriToPath(p.URI)
	key, _ := filepath.Abs(path)
	var diags []Diagnostic
	for _, d := range p.Diagnostics {
		code := ""
		switch v := d.Code.(type) {
		case string:
			code = v
		case float64:
			code = fmt.Sprintf("%.0f", v)
		}
		diags = append(diags, Diagnostic{
			File:      key,
			Line:      d.Range.Start.Line + 1,
			Column:    d.Range.Start.Character + 1,
			EndLine:   d.Range.End.Line + 1,
			EndColumn: d.Range.End.Character + 1,
			Severity:  severityFromLSP(d.Severity),
			Message:   d.Message,
			Source:    d.Source,
			Code:      code,
		})
	}
	if len(diags) == 0 {
		c.diags.Delete(key)
		return
	}
	c.diags.Store(key, diags)
}

func (c *Client) request(method string, params any) (json.RawMessage, error) {
	id := c.nextID.Add(1)
	ch := make(chan json.RawMessage, 1)
	c.pending.Store(id, ch)

	body, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"method":  method,
		"params":  params,
	})
	if err != nil {
		c.pending.Delete(id)
		return nil, err
	}
	if err := c.write(body); err != nil {
		c.pending.Delete(id)
		return nil, err
	}

	select {
	case resp := <-ch:
		var envelope struct {
			Result json.RawMessage `json:"result"`
			Error  *struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal(resp, &envelope); err != nil {
			return nil, err
		}
		if envelope.Error != nil {
			return nil, fmt.Errorf("%s", envelope.Error.Message)
		}
		return envelope.Result, nil
	case <-time.After(requestTimeout):
		c.pending.Delete(id)
		return nil, fmt.Errorf("LSP request %s timed out", method)
	}
}

func (c *Client) notify(method string, params any) error {
	body, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
	})
	if err != nil {
		return err
	}
	return c.write(body)
}

func (c *Client) write(body []byte) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	_, err := c.writer.Write(Encode(body))
	return err
}

func (c *Client) SyncDocument(path, content, languageID string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	uri := pathToURI(abs)

	if docAny, ok := c.openDocs.Load(abs); ok {
		doc := docAny.(openDoc)
		doc.version++
		c.openDocs.Store(abs, doc)
		return c.notify("textDocument/didChange", map[string]any{
			"textDocument": map[string]any{"uri": uri, "version": doc.version},
			"contentChanges": []map[string]any{{"text": content}},
		})
	}

	c.openDocs.Store(abs, openDoc{uri: uri, languageID: languageID, version: 1})
	return c.notify("textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{
			"uri":        uri,
			"languageId": languageID,
			"version":    1,
			"text":       content,
		},
	})
}

func (c *Client) Diagnostics(path string) []Diagnostic {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil
	}
	if v, ok := c.diags.Load(abs); ok {
		return append([]Diagnostic(nil), v.([]Diagnostic)...)
	}
	return nil
}

func (c *Client) AllDiagnostics() []Diagnostic {
	var all []Diagnostic
	c.diags.Range(func(_, value any) bool {
		all = append(all, value.([]Diagnostic)...)
		return true
	})
	return all
}

func (c *Client) Close() {
	c.closeOnce.Do(func() {
		go func() {
			_, _ = c.request("shutdown", nil)
			_ = c.notify("exit", nil)
		}()
		if c.writer != nil {
			_ = c.writer.Close()
		}
		if c.cmd != nil && c.cmd.Process != nil {
			_ = c.cmd.Process.Kill()
		}
		<-c.done
	})
}

func pathToURI(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	u := url.URL{Scheme: "file", Path: filepath.ToSlash(abs)}
	return u.String()
}

func uriToPath(uri string) string {
	u, err := url.Parse(uri)
	if err != nil {
		return uri
	}
	return filepath.FromSlash(u.Path)
}
