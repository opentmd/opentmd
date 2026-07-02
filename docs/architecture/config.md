# 配置系统

## 概述

配置系统位于 `internal/config/`，负责全局配置文件的读写、Provider 预设管理、目录初始化。项目级配置由 `internal/project/` 管理。

## 全局配置路径

```bash
~/.opentmd/config.toml
```

## 默认配置结构

```toml
[default]
provider = "deepseek"
model = "deepseek-chat"

[providers.deepseek]
base_url = "https://api.deepseek.com"
api_key = ""

[providers.openai]
base_url = "https://api.openai.com/v1"
api_key = ""

[session]
persist = true

[tui]
theme = "dark"
stream = true

[agent]
max_iterations = 10

[setup]
completed = false
```

## 核心类型

### `Config` 结构体 (config.go)

```go
type Config struct {
    Default struct {
        Provider string
        Model    string
    }
    Providers map[string]ProviderConfig
    Session   struct { Persist bool }
    TUI       struct { Theme string; Stream bool }
    Agent     struct { MaxIterations int }
    Setup     struct { Completed bool }
}

type ProviderConfig struct {
    BaseURL string
    APIKey  string
}
```

## Provider 预设

| Provider | 标签 | 默认模型 | 需 API Key |
|----------|------|----------|-----------|
| deepseek | DeepSeek | deepseek-chat | ✅ |
| openai | OpenAI | gpt-4o | ✅ |
| claude | Claude (via OpenRouter) | anthropic/claude-3.5-sonnet | ✅ |
| glm | GLM (智谱) | glm-4 | ✅ |
| qwen | Qwen (通义千问) | qwen-max | ✅ |
| ollama | Ollama (本地) | llama3 | ❌ |

所有 Provider 使用 OpenAI 兼容接口，通过 `deepseek` 包统一调用。

## 全局目录布局

首次运行 `opentmd` 时由 `config.Ensure()` 自动创建：

```bash
~/.opentmd/
├── config.toml     # 主配置
├── OPENTMD.md      # 全局指令（可选，注入 system prompt）
├── mcp.json        # MCP 服务器配置
├── hooks.json      # Hooks 配置
├── sessions/       # 会话存储
├── plugins/        # 插件目录
└── skills/         # 全局 Skills（*/SKILL.md）
    └── example/
        └── SKILL.md
```

## 项目目录布局

进入项目运行 `opentmd` 时由 `project.Ensure()` 自动创建：

```bash
<project>/
├── OPENTMD.md               # 项目指令（可选，大小写不敏感）
├── .opentmd.user.md         # 个人指令（gitignore，不提交）
├── .mcp.json                # 项目 MCP（可选，gitignore）
├── .hooks.json              # 项目 Hooks（可选，gitignore）
└── .opentmd/
    ├── skills/              # 项目 Skills
    ├── local/               # 本地缓存（gitignore）
    ├── mcp.json             # 项目 MCP 备选路径
    ├── hooks.json           # 项目 Hooks 备选路径
    └── graph.gob            # Go 代码图谱缓存（自动生成）
```

### 配置加载优先级

| 类型 | 全局 | 项目 |
|------|------|------|
| 指令 | `~/.opentmd/OPENTMD.md`（可选） | `OPENTMD.md` / `CLAUDE.md`（大小写不敏感，可选） |
| Skills | `~/.opentmd/skills/` | `.opentmd/skills/`（含上级目录） |
| MCP | `~/.opentmd/mcp.json` | `.mcp.json` 或 `.opentmd/mcp.json` |
| Hooks | `~/.opentmd/hooks.json` | `.hooks.json` 或 `.opentmd/hooks.json` |

同名 Skill 以**更靠近工作目录**的定义为准。

## 关键函数

| 函数 | 模块 | 功能 |
|------|------|------|
| `Ensure()` | config | 创建 `~/.opentmd` 及默认文件 |
| `Load()` | config | 加载配置文件 |
| `Save()` | config | 保存配置文件 |
| `SetAPIKey()` | config | 设置 Provider API Key |
| `MarkSetupCompleted()` | config | 标记向导完成 |
| `NeedsWizard()` | config | 判断是否需启动向导 |
| `ActiveProvider()` | config | 获取当前活跃 Provider |
| `Ensure(workDir)` | project | 创建项目 `.opentmd/` 布局 |
| `LoadContext(workDir)` | project | 加载全局/项目/用户指令 |
