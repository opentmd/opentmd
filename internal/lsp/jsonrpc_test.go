package lsp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
	"testing"
)

func TestEncodeFormat(t *testing.T) {
	body := []byte(`{"jsonrpc":"2.0"}`)
	encoded := string(Encode(body))
	want := "Content-Length: 17\r\n\r\n" + string(body)
	if encoded != want {
		t.Fatalf("got %q want %q", encoded, want)
	}
}

func TestReadMessageParses(t *testing.T) {
	body := `{"jsonrpc":"2.0","id":1}`
	raw := "Content-Length: " + strconv.Itoa(len(body)) + "\r\n\r\n" + body
	r := bufio.NewReader(bytes.NewReader([]byte(raw)))
	msg, err := ReadMessage(r)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(msg), `"id":1`) {
		t.Fatalf("unexpected msg %s", msg)
	}
}

func TestHandleDiagnostics(t *testing.T) {
	c := &Client{}
	params := json.RawMessage(`{
		"uri":"file:///tmp/test.go",
		"diagnostics":[{
			"range":{"start":{"line":9,"character":4},"end":{"line":9,"character":8}},
			"severity":1,
			"message":"undefined: foo",
			"source":"gopls",
			"code":"UndeclaredName"
		}]
	}`)
	c.handleDiagnostics(params)
	diags := c.Diagnostics("/tmp/test.go")
	if len(diags) != 1 {
		t.Fatalf("expected 1 diag, got %d", len(diags))
	}
	if diags[0].Line != 10 || diags[0].Message != "undefined: foo" {
		t.Fatalf("unexpected diag: %+v", diags[0])
	}
}

func TestDiagnosticDisplayLine(t *testing.T) {
	d := Diagnostic{
		File:     "main.go",
		Line:     3,
		Column:   1,
		Severity: SeverityError,
		Message:  "expected declaration",
		Source:   "gopls",
		Code:     "X",
	}
	got := d.DisplayLine()
	if !strings.Contains(got, "[ERROR]") || !strings.Contains(got, "expected declaration") {
		t.Fatalf("unexpected display: %s", got)
	}
}

func TestFilterDiagnostics(t *testing.T) {
	diags := []Diagnostic{
		{Severity: SeverityError, Message: "e"},
		{Severity: SeverityWarning, Message: "w"},
	}
	errs := FilterDiagnostics(diags, "error")
	if len(errs) != 1 || errs[0].Message != "e" {
		t.Fatalf("filter error failed: %+v", errs)
	}
}
