package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/opentmd/opentmd/internal/config"
	"github.com/opentmd/opentmd/internal/llm"
)

type Message struct {
	Role      llm.Role  `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

type Session struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Messages  []Message `json:"messages"`
}

type Options struct {
	ContinueLast bool
}

type Store struct {
	dir     string
	persist bool
	current *Session
}

func openStore(cfg *config.Config) (*Store, error) {
	dir, err := config.SessionsDir()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create sessions dir: %w", err)
	}
	return &Store{dir: dir, persist: cfg.Session.Persist}, nil
}

func NewStore(cfg *config.Config) (*Store, error) {
	return NewStoreWithOptions(cfg, Options{})
}

func NewStoreWithOptions(cfg *config.Config, opts Options) (*Store, error) {
	s, err := openStore(cfg)
	if err != nil {
		return nil, err
	}
	if opts.ContinueLast {
		if err := s.restoreLatest(); err != nil {
			return nil, err
		}
	} else if err := s.New(); err != nil {
		return nil, err
	}
	return s, nil
}

// NewIsolatedStore creates a store without restoring the latest session (for daemon/API use).
func NewIsolatedStore(cfg *config.Config) (*Store, error) {
	return openStore(cfg)
}

func (s *Store) restoreLatest() error {
	sessions, err := s.List()
	if err != nil {
		return err
	}
	if len(sessions) > 0 {
		latest := sessions[0]
		s.current = &latest
		return nil
	}
	return s.New()
}

func (s *Store) New() error {
	now := time.Now()
	s.current = &Session{
		ID:        uuid.New().String(),
		Title:     "New Session",
		CreatedAt: now,
		UpdatedAt: now,
		Messages:  []Message{},
	}
	return s.saveCurrent()
}

func (s *Store) Current() *Session {
	return s.current
}

func (s *Store) AddMessage(role llm.Role, content string) error {
	if s.current == nil {
		if err := s.New(); err != nil {
			return err
		}
	}
	msg := Message{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	}
	s.current.Messages = append(s.current.Messages, msg)
	s.current.UpdatedAt = time.Now()

	if s.current.Title == "New Session" && role == llm.RoleUser {
		s.current.Title = truncateTitle(content, 40)
	}
	return s.saveCurrent()
}

func truncateTitle(s string, maxRunes int) string {
	if utf8.RuneCountInString(s) <= maxRunes {
		return s
	}
	runes := []rune(s)
	return string(runes[:maxRunes]) + "..."
}

func (s *Store) ProviderMessages() []llm.Message {
	if s.current == nil {
		return nil
	}
	out := make([]llm.Message, len(s.current.Messages))
	for i, m := range s.current.Messages {
		out[i] = llm.Message{Role: m.Role, Content: m.Content}
	}
	return out
}

func (s *Store) Clear() error {
	return s.New()
}

// ReplaceMessages swaps the current session messages and persists.
func (s *Store) ReplaceMessages(messages []Message) error {
	if s.current == nil {
		return fmt.Errorf("no active session")
	}
	s.current.Messages = append([]Message{}, messages...)
	s.current.UpdatedAt = time.Now()
	return s.saveCurrent()
}

func validateSessionID(id string) error {
	if id == "" {
		return fmt.Errorf("session id is empty")
	}
	if strings.Contains(id, "..") || strings.ContainsAny(id, `/\`) {
		return fmt.Errorf("invalid session id")
	}
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("invalid session id format")
	}
	return nil
}

func (s *Store) Load(id string) error {
	if err := validateSessionID(id); err != nil {
		return err
	}
	path := filepath.Join(s.dir, id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("load session %q: %w", id, err)
	}
	var sess Session
	if err := json.Unmarshal(data, &sess); err != nil {
		return fmt.Errorf("parse session: %w", err)
	}
	s.current = &sess
	return nil
}

func (s *Store) List() ([]Session, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, err
	}
	var sessions []Session
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(s.dir, e.Name()))
		if err != nil {
			continue
		}
		var sess Session
		if err := json.Unmarshal(data, &sess); err != nil {
			continue
		}
		sessions = append(sessions, sess)
	}
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].UpdatedAt.After(sessions[j].UpdatedAt)
	})
	return sessions, nil
}

func (s *Store) saveCurrent() error {
	if !s.persist || s.current == nil {
		return nil
	}
	data, err := json.MarshalIndent(s.current, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}
	path := filepath.Join(s.dir, s.current.ID+".json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write session: %w", err)
	}
	return nil
}
