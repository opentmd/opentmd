package project

import (
	"fmt"
	"os"
	"path/filepath"
)

const DirName = ".opentmd"

// LayoutSubdirs are created under <project>/.opentmd on startup.
var LayoutSubdirs = []string{"skills", "local"}

func Dir(workDir string) string {
	return filepath.Join(workDir, DirName)
}

func SkillsDir(workDir string) string {
	return filepath.Join(Dir(workDir), "skills")
}

func LocalDir(workDir string) string {
	return filepath.Join(Dir(workDir), "local")
}

func GraphCachePath(workDir string) string {
	return filepath.Join(Dir(workDir), "graph.gob")
}

// Ensure creates the project .opentmd layout and optional starter files.
func Ensure(workDir string) error {
	root := Dir(workDir)
	if err := os.MkdirAll(root, 0o755); err != nil {
		return fmt.Errorf("create project opentmd dir: %w", err)
	}
	for _, name := range LayoutSubdirs {
		sub := filepath.Join(root, name)
		if err := os.MkdirAll(sub, 0o755); err != nil {
			return fmt.Errorf("create %q: %w", sub, err)
		}
	}
	if err := ensureProjectInstructions(workDir); err != nil {
		return err
	}
	if err := ensureExampleSkill(root); err != nil {
		return err
	}
	if err := ensureProjectMCPConfig(root); err != nil {
		return err
	}
	if err := ensureProjectHooksConfig(root); err != nil {
		return err
	}
	if err := ensureLocalGitkeep(root); err != nil {
		return err
	}
	return nil
}

func ensureProjectMCPConfig(root string) error {
	path := filepath.Join(root, "mcp.json")
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	example := `{
  "mcpServers": {
    "example": {
      "command": "echo",
      "args": ["MCP server placeholder — replace with real command"],
      "disabled": true
    }
  }
}
`
	return os.WriteFile(path, []byte(example), 0o600)
}

func ensureProjectHooksConfig(root string) error {
	path := filepath.Join(root, "hooks.json")
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	example := `{
  "hooks": {
    "pre_bash": {
      "event": "PreToolUse",
      "matcher": "bash",
      "command": "echo '{\"action\":\"allow\"}'",
      "disabled": true,
      "timeout_ms": 5000
    }
  }
}
`
	return os.WriteFile(path, []byte(example), 0o644)
}

func ensureLocalGitkeep(root string) error {
	path := filepath.Join(root, "local", ".gitkeep")
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	return os.WriteFile(path, nil, 0o644)
}

func ensureProjectInstructions(workDir string) error {
	path := filepath.Join(workDir, ".opentmd.md")
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	example := `# Project Instructions

在此填写本项目的 AI 协作约定，例如：

- 技术栈与目录结构
- 代码风格与测试要求
- 常用构建/测试命令
`
	return os.WriteFile(path, []byte(example), 0o644)
}

func ensureExampleSkill(root string) error {
	dir := filepath.Join(root, "skills", "example")
	path := filepath.Join(dir, "SKILL.md")
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	example := `---
name: example
description: "示例 Skill — 展示 SKILL.md 格式"
---

# Example Skill

用户参数: $ARGUMENTS

请按上述参数完成任务。
`
	return os.WriteFile(path, []byte(example), 0o644)
}
