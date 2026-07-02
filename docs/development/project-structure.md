# 项目结构

## 目录布局

```
opentmd-cli/
├── cmd/
│   └── opentmd/
│       └── main.go          # CLI 入口（Cobra 框架）
├── internal/
│   ├── config/              # 配置读写（TOML）
│   │   ├── config.go        # Config 结构体、读写
│   │   ├── config_test.go   # 配置测试
│   │   └── providers.go     # Provider 预设列表
│   ├── llm/                 # LLM 类型与接口
│   │   └── types.go         # Role, ToolCall, Provider 接口
│   ├── provider/            # Provider 工厂
│   │   └── factory.go       # New() 工厂函数
│   ├── deepseek/            # OpenAI 兼容 HTTP 客户端
│   │   ├── deepseek.go      # Client: Chat, Complete, SSE 解析
│   │   └── deepseek_test.go # 客户端测试
│   ├── agent/               # Agent 多步循环编排
│   │   ├── agent.go         # Agent 结构体、Chat 循环
│   │   └── prompt.go        # System prompt
│   ├── tool/                # 工具注册与执行
│   │   └── registry.go      # Registry: read_file, run_shell, list_dir
│   ├── fileio/              # 文件读取
│   │   └── read.go          # Read(), ValidatePath()
│   ├── shell/               # Shell 命令执行
│   │   └── exec.go          # Run(), FormatResult()
│   ├── session/             # 会话持久化
│   │   ├── session.go       # Store, Session, Message
│   │   └── session_test.go  # 会话测试
│   ├── setup/               # 首次启动向导
│   │   └── wizard.go        # RunWizard, RunLogin, RunOAuth
│   └── tui/                 # Bubble Tea 终端界面
│       ├── tui.go           # Model, Update, View
│       └── wrap.go          # 文本换行、消息样式
├── scripts/
│   ├── build.sh             # 编译脚本
│   └── install.sh           # 一键安装脚本
├── mock/
│   └── atomcode/            # AtomCode Rust 原型项目（测试参考）
├── docs/                    # 文档
├── go.mod                   # Go 模块定义
├── go.sum                   # Go 依赖锁
└── LICENSE                  # 开源协议
```

## 依赖关系

```
cmd/opentmd → internal/agent → internal/provider → internal/deepseek
                              → internal/session
                              → internal/tool → internal/fileio
                                              → internal/shell
             internal/config
             internal/setup
             internal/tui
```

## Go 依赖

| 依赖 | 用途 |
|------|------|
| `github.com/spf13/cobra` | CLI 框架 |
| `github.com/charmbracelet/bubbletea` | TUI 框架 |
| `github.com/charmbracelet/bubbles` | TUI 组件（viewport, textarea, spinner） |
| `github.com/charmbracelet/lipgloss` | 终端样式 |
| `github.com/pelletier/go-toml/v2` | TOML 解析 |
| `github.com/google/uuid` | 会话 ID 生成 |
| `golang.org/x/term` | 终端检测 |
