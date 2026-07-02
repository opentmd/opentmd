package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/opentmd/opentmd-cli/internal/agent"
	"github.com/opentmd/opentmd-cli/internal/config"
	"github.com/opentmd/opentmd-cli/internal/diagnostics"
	"github.com/opentmd/opentmd-cli/internal/lsp"
	"github.com/opentmd/opentmd-cli/internal/mcp"
	"github.com/opentmd/opentmd-cli/internal/permission"
	"github.com/opentmd/opentmd-cli/internal/session"
)

const DefaultPort = 13456
const Version = "0.1.0"

type Server struct {
	cfg      *config.Config
	port     int
	workDir  string
	mu       sync.Mutex
	sessions map[string]*sessionEntry
	httpSrv  *http.Server

	permMu      sync.Mutex
	permPending map[string]chan permission.Decision
}

type sessionEntry struct {
	agent *agent.Agent
}

func New(cfg *config.Config, port int) *Server {
	if port <= 0 {
		port = DefaultPort
	}
	wd, _ := os.Getwd()
	return &Server{
		cfg:         cfg,
		port:        port,
		workDir:     wd,
		sessions:    map[string]*sessionEntry{},
		permPending: map[string]chan permission.Decision{},
	}
}

func (s *Server) SetWorkDir(dir string) {
	if dir != "" {
		s.workDir = dir
	}
}

func (s *Server) ListenAndServe() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/chat", s.handleChat)
	mux.HandleFunc("/sessions", s.handleSessions)
	mux.HandleFunc("/sessions/new", s.handleSessionNew)
	mux.HandleFunc("/config", s.handleConfig)
	mux.HandleFunc("/config/reload", s.handleConfigReload)
	mux.HandleFunc("/mcp/reload", s.handleMCPReload)
	mux.HandleFunc("/mcp/status", s.handleMCPStatus)
	mux.HandleFunc("/lsp/status", s.handleLSPStatus)
	mux.HandleFunc("/lsp/reload", s.handleLSPReload)
	mux.HandleFunc("/permission/respond", s.handlePermissionRespond)

	addr := fmt.Sprintf("127.0.0.1:%d", s.port)
	s.httpSrv = &http.Server{Addr: addr, Handler: mux}
	fmt.Fprintf(os.Stderr, "opentmd-daemon listening on http://%s\n", addr)
	err := s.httpSrv.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.closeAllAgents()
	if s.httpSrv == nil {
		return nil
	}
	return s.httpSrv.Shutdown(ctx)
}

func (s *Server) closeAllAgents() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, entry := range s.sessions {
		entry.agent.Shutdown()
	}
	s.sessions = map[string]*sessionEntry{}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{"status": "ok", "version": Version})
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GET only", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, s.configSummary())
}

func (s *Server) handleConfigReload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	cfg, err := s.reloadConfig()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{
		"status":   "ok",
		"provider": cfg.Default.Provider,
		"model":    cfg.Default.Model,
	})
}

func (s *Server) configSummary() map[string]any {
	return map[string]any{
		"provider": s.cfg.Default.Provider,
		"model":    s.cfg.Default.Model,
		"lsp": map[string]any{
			"enabled":         s.cfg.LSP.Enabled,
			"auto_detect":     s.cfg.LSP.AutoDetect,
			"settle_delay_ms": s.cfg.LSP.DiagnosticsSettleDelayMs,
		},
	}
}

func (s *Server) reloadConfig() (*config.Config, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cfg = cfg
	for _, entry := range s.sessions {
		if err := entry.agent.Reload(cfg); err != nil {
			return nil, err
		}
	}
	return cfg, nil
}

func (s *Server) handleLSPStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GET only", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, s.lspStatus())
}

func (s *Server) handleLSPReload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	s.mu.Lock()
	for _, entry := range s.sessions {
		entry.agent.ReloadLSP()
	}
	s.mu.Unlock()
	writeJSON(w, map[string]string{"status": "ok"})
}

func (s *Server) lspStatus() lsp.Status {
	s.mu.Lock()
	var ag *agent.Agent
	for _, entry := range s.sessions {
		ag = entry.agent
		break
	}
	s.mu.Unlock()
	if ag != nil {
		return ag.LSPStatusStruct()
	}

	coll := diagnostics.NewCollector(s.workDir, s.cfg, nil)
	defer coll.Close()
	return coll.Status()
}

func (s *Server) handleSessions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		sess, err := session.NewIsolatedStore(s.cfg)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		list, _ := sess.List()
		writeJSON(w, list)
	default:
		http.Error(w, "GET only", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleSessionNew(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	ag, sid, err := s.createAgent("", s.workDir)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	_ = ag
	writeJSON(w, map[string]string{"session_id": sid})
}

type chatRequest struct {
	Message   string `json:"message"`
	SessionID string `json:"session_id"`
	WorkDir   string `json:"work_dir"`
}

func (s *Server) handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	var req chatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", 500)
		return
	}

	workDir := req.WorkDir
	if workDir == "" {
		workDir = s.workDir
	}

	ag, sessionID, err := s.agentForSession(req.SessionID, workDir)
	if err != nil {
		sendSSE(w, flusher, "error", map[string]string{"message": err.Error()})
		return
	}

	sendSSE(w, flusher, "session", map[string]string{"session_id": sessionID})

	ctx := r.Context()
	ag.SetApprovalHandler(s.sseApprovalHandler(ctx, w, flusher))
	ag.SetLSPEventHandler(func(ev lsp.ConnectEvent) {
		data := map[string]string{
			"command": ev.Command,
			"ext":     ev.Ext,
		}
		if ev.Started {
			data["status"] = "started"
		} else if ev.Error != "" {
			data["status"] = "failed"
			data["error"] = ev.Error
		} else {
			return
		}
		sendSSE(w, flusher, "lsp_connect", data)
	})
	defer ag.SetLSPEventHandler(nil)

	_, err = ag.ChatWithStatus(ctx, req.Message, func(chunk string) error {
		sendSSE(w, flusher, "text", map[string]string{"content": chunk})
		return nil
	}, func(status string) {
		sendSSE(w, flusher, "tool_start", map[string]string{"status": status})
	})
	if err != nil && err != context.Canceled {
		sendSSE(w, flusher, "error", map[string]string{"message": err.Error()})
	}
	sendSSE(w, flusher, "done", map[string]string{
		"status":     "ok",
		"session_id": sessionID,
	})
}

type permissionRespondRequest struct {
	RequestID string `json:"request_id"`
	Decision  string `json:"decision"`
}

func (s *Server) handlePermissionRespond(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	var req permissionRespondRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	decision := parseDecision(req.Decision)
	if !s.deliverPermission(req.RequestID, decision) {
		http.Error(w, "unknown or expired request_id", 404)
		return
	}
	writeJSON(w, map[string]string{"status": "ok"})
}

func (s *Server) handleMCPReload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	ctx := r.Context()
	for _, entry := range s.sessions {
		_ = entry.agent.ReloadMCP(ctx)
	}
	writeJSON(w, map[string]string{"status": "ok"})
}

func (s *Server) handleMCPStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GET only", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, s.mcpStatus(r.Context()))
}

func (s *Server) mcpStatus(ctx context.Context) mcp.StatusReport {
	s.mu.Lock()
	var ag *agent.Agent
	for _, entry := range s.sessions {
		ag = entry.agent
		break
	}
	s.mu.Unlock()
	if ag != nil {
		return ag.MCPStatus(ctx)
	}
	return mcp.ProbeStatus(ctx, s.workDir)
}

func (s *Server) agentForSession(sessionID, workDir string) (*agent.Agent, string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if sessionID != "" {
		if entry, ok := s.sessions[sessionID]; ok {
			if workDir != "" {
				_ = entry.agent.SetWorkDir(workDir)
			}
			return entry.agent, sessionID, nil
		}
	}

	return s.createAgentLocked(sessionID, workDir)
}

func (s *Server) createAgent(sessionID, workDir string) (*agent.Agent, string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.createAgentLocked(sessionID, workDir)
}

func (s *Server) createAgentLocked(sessionID, workDir string) (*agent.Agent, string, error) {
	store, err := session.NewIsolatedStore(s.cfg)
	if err != nil {
		return nil, "", err
	}
	if sessionID != "" {
		if err := store.Load(sessionID); err != nil {
			return nil, "", err
		}
	} else if err := store.New(); err != nil {
		return nil, "", err
	}

	if workDir == "" {
		workDir = s.workDir
	}

	ag, err := agent.New(s.cfg, store, agent.Options{WorkDir: workDir})
	if err != nil {
		return nil, "", err
	}

	sid := store.Current().ID
	s.sessions[sid] = &sessionEntry{agent: ag}
	return ag, sid, nil
}

func (s *Server) sseApprovalHandler(ctx context.Context, w http.ResponseWriter, flusher http.Flusher) permission.Handler {
	return func(req permission.Request) permission.Decision {
		reqID := uuid.New().String()
		ch := s.registerPermission(reqID)
		defer s.unregisterPermission(reqID)

		sendSSE(w, flusher, "permission_request", map[string]string{
			"request_id": reqID,
			"tool":       req.Tool,
			"args":       req.Args,
			"reason":     req.Reason,
		})

		select {
		case d := <-ch:
			return d
		case <-ctx.Done():
			return permission.Deny
		case <-time.After(5 * time.Minute):
			return permission.Deny
		}
	}
}

func (s *Server) registerPermission(id string) chan permission.Decision {
	ch := make(chan permission.Decision, 1)
	s.permMu.Lock()
	s.permPending[id] = ch
	s.permMu.Unlock()
	return ch
}

func (s *Server) unregisterPermission(id string) {
	s.permMu.Lock()
	delete(s.permPending, id)
	s.permMu.Unlock()
}

func (s *Server) deliverPermission(id string, d permission.Decision) bool {
	s.permMu.Lock()
	ch, ok := s.permPending[id]
	s.permMu.Unlock()
	if !ok {
		return false
	}
	select {
	case ch <- d:
		return true
	default:
		return false
	}
}

func parseDecision(s string) permission.Decision {
	switch s {
	case "allow_once", "y", "once":
		return permission.AllowOnce
	case "allow_session", "a", "session":
		return permission.AllowSession
	default:
		return permission.Deny
	}
}

func sendSSE(w http.ResponseWriter, flusher http.Flusher, event string, data any) {
	b, _ := json.Marshal(data)
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, b)
	flusher.Flush()
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
