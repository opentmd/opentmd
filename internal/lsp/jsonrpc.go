package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func Encode(body []byte) []byte {
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(body))
	msg := make([]byte, 0, len(header)+len(body))
	msg = append(msg, header...)
	msg = append(msg, body...)
	return msg
}

func ReadMessage(r *bufio.Reader) (json.RawMessage, error) {
	contentLength := 0
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("read LSP header: %w", err)
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		if strings.HasPrefix(strings.ToLower(line), "content-length:") {
			n, err := strconv.Atoi(strings.TrimSpace(line[len("Content-Length:"):]))
			if err != nil {
				return nil, fmt.Errorf("parse Content-Length: %w", err)
			}
			contentLength = n
		}
	}
	if contentLength <= 0 {
		return nil, fmt.Errorf("missing Content-Length in LSP message")
	}
	body := make([]byte, contentLength)
	if _, err := io.ReadFull(r, body); err != nil {
		return nil, fmt.Errorf("read LSP body: %w", err)
	}
	return json.RawMessage(body), nil
}
