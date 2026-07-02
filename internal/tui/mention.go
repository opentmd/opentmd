package tui

import (
	"io/fs"
	"path/filepath"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

const mentionMenuMaxVisible = 4

type fileEntry struct {
	relPath string
	isDir   bool
	depth   int
}

type fileIndex struct {
	root    string
	entries []fileEntry
	built   bool
}

func newFileIndex(root string) *fileIndex {
	return &fileIndex{root: root}
}

func (fi *fileIndex) ensure() {
	if fi.built {
		return
	}
	fi.built = true
	fi.entries = fi.walk()
}

func (fi *fileIndex) walk() []fileEntry {
	var out []fileEntry
	_ = filepath.WalkDir(fi.root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, err := filepath.Rel(fi.root, path)
		if err != nil || rel == "." {
			return nil
		}
		if shouldSkipMentionPath(rel) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		isDir := d.IsDir()
		s := filepath.ToSlash(rel)
		if isDir {
			s += "/"
		}
		depth := strings.Count(s, "/")
		if isDir {
			depth++
		}
		out = append(out, fileEntry{relPath: s, isDir: isDir, depth: depth})
		return nil
	})
	return out
}

func shouldSkipMentionPath(rel string) bool {
	parts := strings.Split(filepath.ToSlash(rel), "/")
	for _, p := range parts {
		switch p {
		case ".git", "node_modules", "vendor", ".opentmd", "target", "dist", "build":
			return true
		}
		if strings.HasPrefix(p, ".") && p != "." {
			return true
		}
	}
	return false
}

func (fi *fileIndex) filter(scopeDir, filter string) []fileEntry {
	fi.ensure()
	filterLower := strings.ToLower(filter)
	scopeDepth := strings.Count(scopeDir, "/")
	if scopeDir != "" {
		scopeDepth++
	}

	var matched []fileEntry
	for _, e := range fi.entries {
		if !strings.HasPrefix(e.relPath, scopeDir) {
			continue
		}
		if e.relPath == scopeDir {
			continue
		}
		if filterLower == "" {
			if e.depth != scopeDepth+1 {
				continue
			}
		} else {
			after := e.relPath[len(scopeDir):]
			if !strings.Contains(strings.ToLower(after), filterLower) {
				continue
			}
		}
		matched = append(matched, e)
	}

	for i := 0; i < len(matched); i++ {
		for j := i + 1; j < len(matched); j++ {
			a, b := matched[i], matched[j]
			aDirect := a.depth == scopeDepth+1
			bDirect := b.depth == scopeDepth+1
			swap := false
			if aDirect != bDirect {
				swap = bDirect
			} else if a.isDir != b.isDir {
				swap = b.isDir
			} else {
				swap = a.relPath > b.relPath
			}
			if swap {
				matched[i], matched[j] = matched[j], matched[i]
			}
		}
	}
	if len(matched) > 30 {
		matched = matched[:30]
	}
	return matched
}

func detectAtMentionRange(buf string, cursor int) (int, int, bool) {
	if cursor < 0 || cursor > len(buf) {
		return 0, 0, false
	}
	prefix := buf[:cursor]
	atPos := strings.LastIndex(prefix, "@")
	if atPos < 0 {
		return 0, 0, false
	}
	if atPos > 0 {
		before := rune(prefix[atPos-1])
		if !unicode.IsSpace(before) {
			return 0, 0, false
		}
	}
	tokenToCursor := prefix[atPos+1:]
	for _, r := range tokenToCursor {
		if unicode.IsSpace(r) {
			return 0, 0, false
		}
	}
	afterAt := buf[atPos+1:]
	end := len(afterAt)
	for i, r := range afterAt {
		if unicode.IsSpace(r) {
			end = i
			break
		}
	}
	return atPos, atPos + 1 + end, true
}

func splitMentionToken(token string) (scopeDir, filter string) {
	if i := strings.LastIndex(token, "/"); i >= 0 {
		return token[:i+1], token[i+1:]
	}
	return "", token
}

func (m *Model) inputCursorByte() int {
	val := m.input.Value()
	row := m.input.Line()
	lines := strings.Split(val, "\n")
	if row >= len(lines) {
		return len(val)
	}
	info := m.input.LineInfo()
	col := info.StartColumn + info.ColumnOffset
	lineRunes := []rune(lines[row])
	if col > len(lineRunes) {
		col = len(lineRunes)
	}
	offset := 0
	for i := 0; i < row; i++ {
		offset += len(lines[i]) + 1
	}
	return offset + len(string(lineRunes[:col]))
}

func (m *Model) mentionQuery() (token string, ok bool) {
	buf := m.input.Value()
	cursor := m.inputCursorByte()
	atPos, end, found := detectAtMentionRange(buf, cursor)
	if !found {
		return "", false
	}
	return buf[atPos+1 : end], true
}

func (m *Model) mentionMenuVisible() bool {
	if m.trustPending || m.approvalPending != nil || m.modalActive() {
		return false
	}
	if strings.HasPrefix(strings.TrimSpace(m.slashFirstLine()), "/") {
		return false
	}
	_, ok := m.mentionQuery()
	return ok
}

func (m *Model) filteredMentionEntries() []fileEntry {
	token, ok := m.mentionQuery()
	if !ok {
		return nil
	}
	if m.fileIndex == nil {
		m.fileIndex = newFileIndex(m.agent.WorkDir())
	} else if m.fileIndex.root != m.agent.WorkDir() {
		m.fileIndex = newFileIndex(m.agent.WorkDir())
	}
	scope, filter := splitMentionToken(token)
	return m.fileIndex.filter(scope, filter)
}

func (m *Model) mentionMenuLineCount() int {
	if !m.mentionMenuVisible() {
		return 0
	}
	items := m.filteredMentionEntries()
	if len(items) == 0 {
		return 2
	}
	visible := len(items)
	if visible > mentionMenuMaxVisible {
		visible = mentionMenuMaxVisible
	}
	lines := visible + 1
	if len(items) > mentionMenuMaxVisible {
		lines++
	}
	return lines
}

func (m *Model) syncMentionMenu() {
	if !m.mentionMenuVisible() {
		m.mentionLastQuery = ""
		m.mentionSelected = 0
		m.mentionScroll = 0
		return
	}
	query, _ := m.mentionQuery()
	if query != m.mentionLastQuery {
		m.mentionSelected = 0
		m.mentionScroll = 0
		m.mentionLastQuery = query
	}
	items := m.filteredMentionEntries()
	m.mentionSelected = clampSlashSelection(m.mentionSelected, len(items))
}

func (m *Model) handleMentionMenuKey(msg tea.KeyMsg) (bool, tea.Model, tea.Cmd) {
	if !m.mentionMenuVisible() {
		return false, m, nil
	}
	items := m.filteredMentionEntries()
	if len(items) == 0 {
		return false, m, nil
	}
	switch msg.Type {
	case tea.KeyUp, tea.KeyDown, tea.KeyCtrlP, tea.KeyCtrlN:
		up := msg.Type == tea.KeyUp || msg.Type == tea.KeyCtrlP
		m.mentionSelected = cycleSlashSelection(m.mentionSelected, len(items), up)
		_, _, m.mentionScroll = slashMenuWindow(m.mentionSelected, len(items), mentionMenuMaxVisible, m.mentionScroll)
		return true, m, nil
	case tea.KeyTab:
		m.applyMentionCompletion(items)
		return true, m, nil
	case tea.KeyEnter:
		if msg.Alt {
			return false, m, nil
		}
		if m.shouldUseMentionSelection(items) {
			m.applyMentionChoice(items[clampSlashSelection(m.mentionSelected, len(items))])
			return true, m, nil
		}
	}
	return false, m, nil
}

func (m *Model) shouldUseMentionSelection(items []fileEntry) bool {
	token, ok := m.mentionQuery()
	if !ok {
		return false
	}
	for _, e := range items {
		if e.relPath == token || strings.TrimSuffix(e.relPath, "/") == token {
			return false
		}
	}
	return true
}

func (m *Model) applyMentionCompletion(items []fileEntry) {
	if len(items) == 0 {
		return
	}
	sel := items[clampSlashSelection(m.mentionSelected, len(items))]
	m.applyMentionChoice(sel)
}

func (m *Model) applyMentionChoice(entry fileEntry) {
	buf := m.input.Value()
	cursor := m.inputCursorByte()
	atPos, end, ok := detectAtMentionRange(buf, cursor)
	if !ok {
		return
	}
	path := entry.relPath
	if entry.isDir && !strings.HasSuffix(path, "/") {
		path += "/"
	}
	replacement := "@" + path + " "
	newVal := buf[:atPos] + replacement + buf[end:]
	m.input.SetValue(newVal)
	m.mentionLastQuery = path
}

func (m *Model) renderMentionMenu() string {
	if !m.mentionMenuVisible() {
		return ""
	}
	items := m.filteredMentionEntries()
	if len(items) == 0 {
		return m.styles.hint.Render("  无匹配文件 · Tab 补全 · ↑↓ 选择")
	}

	selected := clampSlashSelection(m.mentionSelected, len(items))
	start, end, scroll := slashMenuWindow(selected, len(items), mentionMenuMaxVisible, m.mentionScroll)
	m.mentionScroll = scroll

	active := m.styles.menuActive
	muted := m.styles.menuInactive

	var lines []string
	for i := start; i < end; i++ {
		item := items[i]
		label := item.relPath
		if item.isDir && !strings.HasSuffix(label, "/") {
			label += "/"
		}
		prefix := "+ "
		if i == selected {
			lines = append(lines, active.Render(prefix+label))
		} else {
			lines = append(lines, muted.Render(prefix+label))
		}
	}
	if len(items) > mentionMenuMaxVisible {
		if end < len(items) {
			lines = append(lines, m.styles.hint.Render("  ↓ more below"))
		} else if start > 0 {
			lines = append(lines, m.styles.hint.Render("  ↑ more above"))
		}
	}

	labelWidth := 0
	for i := start; i < end; i++ {
		if w := runewidth.StringWidth(items[i].relPath); w > labelWidth {
			labelWidth = w
		}
	}

	hint := m.styles.inputBar.Render("Tab 补全 · Enter 选择 · ↑↓ 循环")
	return lipgloss.JoinVertical(lipgloss.Left, strings.Join(lines, "\n"), hint)
}

func (m *Model) invalidateFileIndex() {
	m.fileIndex = nil
}
