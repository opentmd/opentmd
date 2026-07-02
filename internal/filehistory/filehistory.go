package filehistory

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/opentmd/opentmd-cli/internal/config"
)

const maxVersionsPerFile = 50

type fileVersions struct {
	next    uint32
	backups []string
}

// Session tracks per-turn file backups for /undo.
type Session struct {
	sessionID string
	backupDir string
	files     map[string]*fileVersions
	turnFiles []string
}

func NewSession(sessionID string) (*Session, error) {
	root, err := config.Dir()
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(root, "file-history", sessionID)
	return &Session{
		sessionID: sessionID,
		backupDir: dir,
		files:     map[string]*fileVersions{},
	}, nil
}

func (s *Session) BeginTurn() {
	s.turnFiles = nil
}

func (s *Session) Backup(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	if err := os.MkdirAll(s.backupDir, 0o755); err != nil {
		return err
	}
	fv := s.files[path]
	if fv == nil {
		fv = &fileVersions{next: 1}
		s.files[path] = fv
	}
	name := backupName(path, fv.next)
	dest := filepath.Join(s.backupDir, name)
	if err := copyFile(path, dest); err != nil {
		return err
	}
	fv.backups = append(fv.backups, name)
	fv.next++
	for len(fv.backups) > maxVersionsPerFile {
		old := filepath.Join(s.backupDir, fv.backups[0])
		_ = os.Remove(old)
		fv.backups = fv.backups[1:]
	}
	s.turnFiles = appendUnique(s.turnFiles, path)
	return nil
}

func (s *Session) UndoLastTurn() ([]string, error) {
	if len(s.turnFiles) == 0 {
		return nil, fmt.Errorf("no file changes to undo")
	}
	var restored []string
	for _, path := range s.turnFiles {
		fv := s.files[path]
		if fv == nil || len(fv.backups) == 0 {
			continue
		}
		last := fv.backups[len(fv.backups)-1]
		src := filepath.Join(s.backupDir, last)
		if err := copyFile(src, path); err != nil {
			return restored, fmt.Errorf("restore %s: %w", path, err)
		}
		restored = append(restored, path)
		fv.backups = fv.backups[:len(fv.backups)-1]
	}
	s.turnFiles = nil
	if len(restored) == 0 {
		return nil, fmt.Errorf("no backups found for last turn")
	}
	return restored, nil
}

func backupName(path string, version uint32) string {
	sum := sha256.Sum256([]byte(path))
	return hex.EncodeToString(sum[:8]) + fmt.Sprintf("@v%d", version)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

func appendUnique(list []string, item string) []string {
	for _, v := range list {
		if v == item {
			return list
		}
	}
	return append(list, item)
}

func (s *Session) LastTurnFiles() []string {
	out := make([]string, len(s.turnFiles))
	copy(out, s.turnFiles)
	return out
}

func RelDisplay(workDir, path string) string {
	if rel, err := filepath.Rel(workDir, path); err == nil && !strings.HasPrefix(rel, "..") {
		return rel
	}
	return path
}
