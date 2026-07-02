package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/opentmd/opentmd-cli/internal/config"
)

type Skill struct {
	Name        string
	Description string
	Content     string
	Dir         string
	UserInvoke  bool
	LLMInvoke   bool
}

type Registry struct {
	skills map[string]Skill
}

func Load(workDir string) (*Registry, error) {
	r := &Registry{skills: map[string]Skill{}}
	dirs := []string{}
	if home, err := config.Dir(); err == nil {
		dirs = append(dirs, filepath.Join(home, "skills"))
	}
	dirs = append(dirs, projectSkillDirs(workDir)...)
	for _, dir := range dirs {
		_ = scanSkills(dir, r)
	}
	return r, nil
}

func projectSkillDirs(workDir string) []string {
	var chain []string
	for dir := workDir; dir != filepath.Dir(dir); dir = filepath.Dir(dir) {
		chain = append(chain, dir)
	}
	for i, j := 0, len(chain)-1; i < j; i, j = i+1, j-1 {
		chain[i], chain[j] = chain[j], chain[i]
	}
	var dirs []string
	for _, dir := range chain {
		dirs = append(dirs,
			filepath.Join(dir, ".opentmd", "skills"),
			filepath.Join(dir, ".atomcode", "skills"),
		)
	}
	return dirs
}

func scanSkills(root string, r *Registry) error {
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if d.Name() != "SKILL.md" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		s := parseSkill(string(data), filepath.Dir(path))
		if s.Name != "" {
			r.skills[s.Name] = s
		}
		return nil
	})
}

func parseSkill(raw, dir string) Skill {
	body := raw
	name := filepath.Base(dir)
	desc := ""
	if strings.HasPrefix(raw, "---") {
		parts := strings.SplitN(raw, "---", 3)
		if len(parts) >= 3 {
			for _, line := range strings.Split(parts[1], "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "name:") {
					name = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
				}
				if strings.HasPrefix(line, "description:") {
					desc = strings.Trim(strings.TrimSpace(strings.TrimPrefix(line, "description:")), "\"")
				}
			}
			body = strings.TrimSpace(parts[2])
		}
	}
	return Skill{Name: name, Description: desc, Content: body, Dir: dir, UserInvoke: true, LLMInvoke: true}
}

func (r *Registry) Get(name string) (Skill, bool) {
	s, ok := r.skills[name]
	return s, ok
}

func (r *Registry) List() []Skill {
	out := make([]Skill, 0, len(r.skills))
	for _, s := range r.skills {
		out = append(out, s)
	}
	return out
}

func (r *Registry) Expand(name, args string) (string, error) {
	s, ok := r.Get(name)
	if !ok {
		return "", fmt.Errorf("skill %q not found", name)
	}
	out := s.Content
	out = strings.ReplaceAll(out, "$ARGUMENTS", args)
	return out, nil
}

func (r *Registry) PromptSection() string {
	list := r.List()
	if len(list) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("Available Skills (use use_skill tool or /skills <name>):\n")
	for _, s := range list {
		sb.WriteString(fmt.Sprintf("- %s: %s\n", s.Name, s.Description))
	}
	return sb.String()
}
