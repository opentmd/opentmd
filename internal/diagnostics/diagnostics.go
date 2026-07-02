package diagnostics

import (
	"context"

	"github.com/opentmd/opentmd/internal/config"
	"github.com/opentmd/opentmd/internal/lsp"
)

// Collector wraps the LSP manager for the diagnostics tool.
type Collector struct {
	mgr *lsp.Manager
}

func NewCollector(workDir string, cfg *config.Config, onEvent lsp.EventHandler) *Collector {
	return &Collector{mgr: lsp.NewManager(workDir, lsp.ConfigFromApp(cfg), onEvent)}
}

func (c *Collector) Collect(ctx context.Context, filePath, severity string) (string, error) {
	if c.mgr == nil {
		return "LSP not available.", nil
	}
	return c.mgr.Collect(ctx, filePath, severity)
}

func (c *Collector) NotifyFileChanged(ctx context.Context, path, content string) {
	if c.mgr == nil {
		return
	}
	_, _ = c.mgr.NotifyFileChanged(ctx, path, content)
}

func (c *Collector) SetEventHandler(h lsp.EventHandler) {
	if c.mgr != nil {
		c.mgr.SetEventHandler(h)
	}
}

func (c *Collector) Close() {
	if c.mgr != nil {
		c.mgr.Close()
	}
}

func (c *Collector) StatusText() string {
	return c.Status().Text
}

func (c *Collector) Status() lsp.Status {
	if c.mgr == nil {
		return lsp.Status{Text: "LSP: not initialized"}
	}
	return c.mgr.Status()
}
