package lsp

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/opentmd/opentmd/internal/pathutil"
)

type Manager struct {
	workDir string
	cfg     Config
	reg     *Registry
	onEvent EventHandler

	mu      sync.Mutex
	clients map[string]*Client
}

func NewManager(workDir string, cfg Config, onEvent EventHandler) *Manager {
	if cfg.SettleDelay <= 0 {
		cfg.SettleDelay = 300 * time.Millisecond
	}
	if cfg.WarmupFileLimit <= 0 {
		cfg.WarmupFileLimit = 40
	}
	return &Manager{
		workDir: workDir,
		cfg:     cfg,
		reg:     NewRegistry(cfg.AutoDetect, cfg.Servers),
		onEvent: onEvent,
		clients: map[string]*Client{},
	}
}

func (m *Manager) Enabled() bool {
	return m.cfg.Enabled
}

func (m *Manager) EnsureServer(ctx context.Context, filePath string) (*Client, bool, error) {
	if !m.cfg.Enabled {
		return nil, false, nil
	}
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(filePath)), ".")
	if ext == "" {
		return nil, false, nil
	}

	m.mu.Lock()
	if c, ok := m.clients[ext]; ok {
		m.mu.Unlock()
		return c, true, nil
	}
	m.mu.Unlock()

	cfg, ok := m.reg.Get(ext)
	if !ok || !commandAvailable(cfg.Command) {
		return nil, false, nil
	}

	client, err := StartClient(ctx, cfg, m.workDir, languageID(ext))
	if err != nil {
		m.emit(ConnectEvent{Command: cfg.Command, Ext: ext, Error: err.Error()})
		return nil, false, err
	}

	m.mu.Lock()
	if existing, ok := m.clients[ext]; ok {
		m.mu.Unlock()
		client.Close()
		return existing, true, nil
	}
	m.clients[ext] = client
	m.mu.Unlock()
	m.emit(ConnectEvent{Started: true, Command: cfg.Command, Ext: ext})
	return client, true, nil
}

func (m *Manager) NotifyFileChanged(ctx context.Context, filePath, content string) (bool, error) {
	client, ok, err := m.EnsureServer(ctx, filePath)
	if err != nil || !ok {
		return false, err
	}
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(filePath)), ".")
	if err := client.SyncDocument(filePath, content, languageID(ext)); err != nil {
		return false, err
	}
	return true, nil
}

func (m *Manager) SyncFileFromDisk(ctx context.Context, relOrAbs string) (bool, error) {
	path := relOrAbs
	if !filepath.IsAbs(path) {
		path = filepath.Join(m.workDir, relOrAbs)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	return m.NotifyFileChanged(ctx, path, string(data))
}

func (m *Manager) Collect(ctx context.Context, filePath, severity string) (string, error) {
	if !m.cfg.Enabled {
		return "LSP is disabled.", nil
	}
	if severity == "" {
		severity = "error"
	}

	if filePath != "" {
		return m.collectFile(ctx, filePath, severity)
	}
	return m.collectProject(ctx, severity)
}

func (m *Manager) collectFile(ctx context.Context, filePath, severity string) (string, error) {
	path := filePath
	if !filepath.IsAbs(path) {
		path = filepath.Join(m.workDir, filePath)
	}
	ok, err := m.SyncFileFromDisk(ctx, path)
	if err != nil {
		return "", err
	}
	if !ok {
		return fmt.Sprintf("No LSP server available for %s (install language server or check PATH).", filepath.Ext(path)), nil
	}
	time.Sleep(m.cfg.SettleDelay)

	client, running, _ := m.EnsureServer(ctx, path)
	if !running || client == nil {
		return "LSP server not running.", nil
	}
	diags := client.Diagnostics(path)
	filtered := FilterDiagnostics(diags, severity)
	if len(filtered) == 0 {
		return fmt.Sprintf("No diagnostics found for %s (filter: %s).", filePath, severity), nil
	}
	return fmt.Sprintf("%d diagnostic(s) for %s:\n%s", len(filtered), filePath, FormatDiagnostics(filtered, "all")), nil
}

func (m *Manager) collectProject(ctx context.Context, severity string) (string, error) {
	synced, err := m.warmupProject(ctx)
	if err != nil {
		return "", err
	}
	if synced == 0 {
		return "No LSP servers available for this project. Install gopls, rust-analyzer, typescript-language-server, or pylsp.", nil
	}
	time.Sleep(m.cfg.SettleDelay)

	var all []Diagnostic
	m.mu.Lock()
	for _, client := range m.clients {
		all = append(all, client.AllDiagnostics()...)
	}
	m.mu.Unlock()

	filtered := FilterDiagnostics(all, severity)
	if len(filtered) == 0 {
		return fmt.Sprintf("No diagnostics found (filter: %s). Synced %d file(s).", severity, synced), nil
	}
	return fmt.Sprintf("%d diagnostic(s) across project (synced %d files):\n%s",
		len(filtered), synced, FormatDiagnostics(filtered, "all")), nil
}

func (m *Manager) warmupProject(ctx context.Context) (int, error) {
	count := 0
	_ = filepath.WalkDir(m.workDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			if d != nil && d.IsDir() && path != m.workDir && pathutil.ShouldSkipDir(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if count >= m.cfg.WarmupFileLimit {
			return fs.SkipAll
		}
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".go", ".rs", ".ts", ".tsx", ".js", ".jsx", ".py", ".java":
		default:
			return nil
		}
		if _, ok := m.reg.ForFile(path); !ok {
			return nil
		}
		ok, err := m.SyncFileFromDisk(ctx, path)
		if err != nil {
			return nil
		}
		if ok {
			count++
		}
		return nil
	})
	return count, nil
}

func (m *Manager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, c := range m.clients {
		c.Close()
	}
	m.clients = map[string]*Client{}
}

func (m *Manager) SetEventHandler(h EventHandler) {
	m.onEvent = h
}

func (m *Manager) emit(ev ConnectEvent) {
	if m.onEvent != nil {
		m.onEvent(ev)
	}
}
