# TUI 终端界面

## 概述

TUI 位于 `internal/tui/`，基于 [Bubble Tea](https://github.com/charmbracelet/bubbletea) 框架实现，提供终端交互界面。支持 Linux、macOS 以及 Windows（CMD.exe、PowerShell、Windows Terminal）。

## 技术栈

| 组件 | 用途 |
|------|------|
| `bubbletea` | TUI 框架 (Elm 架构) |
| `bubbles` | 组件库 (viewport, textarea, spinner) |
| `lipgloss` | 样式渲染 |
| `lipgloss/x/ansi` | ANSI 处理 |
| `go-runewidth` | 中日韩字符宽度 |

## 界面结构

```
┌─────────────────────────────────────────────────────┐
│   ___  ____  _____  _   _    _____ __  __ ____      │  ← ASCII 艺术 Logo
│  / _ \|  _ \| ____|| \ | |  |_   _||  \/  ||  _ \  │     (终端宽度≥50列时显示)
│ ...                                                  │
│  ◆ OpenTMD  0.1.0  ·  MIT                           │
│  ∙ deepseek/deepseek-chat  · ~/my-project            │
│  描述任务… (/help /provider /login)                  │
│                                                     │
│  ┌─────────────────────────────────────────────────┐ │
│  │  user >  你好                                   │ │  ← Viewport (滚动消息区)
│  │  assistant >  你好！我能帮你做什么？             │ │
│  └─────────────────────────────────────────────────┘ │
│  ────────────────────────────────────────────────── │
│  ❯ 输入框...                                        │  ← 输入区
│  ────────────────────────────────────────────────── │
│  deepseek/deepseek-chat · ~/my-project   Ctrl+Y 复制│  ← 状态栏
└─────────────────────────────────────────────────────┘
```

启用 `--dangerously-skip-permissions` 时，状态栏左侧显示红色徽章：

```
 ⚠ BYPASS  deepseek/...  · ~/my-project          ⠋ 流式输出
```

## 核心文件

| 文件 | 职责 |
|------|------|
| `tui.go` | Model 定义、Update 循环、Run() 入口 |
| `render.go` | 状态栏、输入区、欢迎屏渲染 |
| `wrap.go` | 消息渲染、文本换行、权限审批弹窗 |
| `theme.go` | 主题色 (dark/light)、样式集合 |
| `symbols.go` | 平台自适应符号（Unicode / ASCII 回退） |
| `terminal.go` | 终端能力检测 (`Capabilities`) |
| `terminal_windows.go` | Windows 控制台初始化（仅 Windows 编译） |
| `console_windows.go` | UTF-8 代码页 + VT 处理开关（init() 自动执行） |
| `screen.go` | 备用屏幕策略 (`useAltScreen`) |
| `keys.go` | 快捷键处理与文档 |
| `clipboard.go` | 多后端剪贴板（wl-copy/xclip/OSC52/atotto） |

## 核心组件

### Model (tui.go)

Bubble Tea Model，管理：

| 字段 | 说明 |
|------|------|
| `agent` | Agent 实例引用 |
| `viewport` | 消息视口 (滚动) |
| `input` | 输入框 (textarea) |
| `messages` | 显示消息列表 |
| `spinnerFrame` | 动画帧计数 |
| `autoRun` | 自动运行模式标志 |

### 消息类型

TUI 内部通过 `tea.Msg` 传递异步事件：

| 消息 | 说明 |
|------|------|
| `chunkMsg` | 流式输出 chunk |
| `statusMsg` | 工具状态更新 |
| `doneMsg` | 流式输出完成 |
| `approvalMsg` | 权限审批请求 |
| `lspConnectMsg` | LSP 连接事件 |

## ASCII 艺术 Logo

进入 TUI 且无消息时，欢迎屏居中显示 "OPEN TMD" ASCII 艺术 Logo（使用 figlet standard 字体风格，纯 ASCII 字符），在所有平台均可渲染：

```
  ___  ____  _____  _   _    _____ __  __ ____
 / _ \|  _ \| ____|| \ | |  |_   _||  \/  ||  _ \
| | | || |_) |  _|  |  \| |    | | | |\/| || | | |
| |_| ||  __/| |___ | |\  |    | | | |  | || |_| |
 \___/ |_|   |_____|_| \_|    |_| |_|  |_||____/
```

终端宽度不足 50 列时退化为紧凑模式（仅显示 `◆ OpenTMD`）。

## 平台自适应符号 (symbols.go)

```go
type termSymbols struct {
    Spinner  []string   // 动画帧
    Prompt   string     // 输入提示符
    Bullet   string     // 列表项
    Diamond  string     // 品牌图标
    // ... 更多
}
```

| 场景 | 符号示例 |
|------|---------|
| Unicode 终端（Linux/macOS/Windows Terminal） | `❯ ` `∙ ` `◆ ` `⠋⠙⠹…` |
| ASCII 终端（Windows CMD.exe 经典模式等） | `> ` `* ` `* ` `\|/-\` |

`asciiMode()` 判断逻辑：

1. `OPENTMD_ASCII=1` → 强制 ASCII
2. `TERM=dumb` → ASCII
3. 非 Windows → Unicode
4. Windows Terminal (`WT_SESSION`)、ConEmu (`ConEmuANSI=ON`)、VS Code (`TERM_PROGRAM`)、Git Bash (`MSYSTEM`) → Unicode
5. 经典 CMD.exe / 旧 PowerShell → ASCII

## Windows 控制台支持 (console_windows.go)

程序启动时（`init()`）自动执行：

1. `SetConsoleCP(65001)` + `SetConsoleOutputCP(65001)` — 切换到 UTF-8 代码页
2. `SetConsoleMode` 添加 `ENABLE_VIRTUAL_TERMINAL_PROCESSING` — 启用 ANSI 颜色与样式
3. `SetConsoleMode` 添加 `DISABLE_NEWLINE_AUTO_RETURN` — 防止布局错乱

仅在 Windows 构建（`//go:build windows`）中编译。

## 文字复制

TUI 使用 `WithMouseCellMotion()`（xterm 1002 模式）而非 `WithMouseAllMotion()`，鼠标滚轮滚动正常工作，同时在主流终端中支持 `Shift+拖拽` 进行原生文字选中。

复制方式汇总：

| 方式 | 说明 |
|------|------|
| `Ctrl+Y` | 复制最近 AI 回复到系统剪贴板 |
| `/copy` | 同 `Ctrl+Y` |
| `/copy all` | 复制整个会话 |
| `Shift+拖拽` | 终端原生选中（主流终端均支持） |
| `OPENTMD_NO_ALT_SCREEN=1` | 禁用备用屏幕，可使用终端滚动缓冲区选中 |

复制失败时自动写入 `~/.opentmd/last-copy.txt`。

状态栏在有 AI 回复且视口在底部时显示 `Ctrl+Y 复制` 提示。

## Slash 命令

| 命令 | 说明 |
|------|------|
| `/help` | 显示全部命令与快捷键 |
| `/clear` | 清空当前会话显示 |
| `/session list/new/<id>` | 会话管理 |
| `/copy [all]` | 复制 AI 回复 / 完整会话 |
| `/export [file]` | 导出会话到文件 |
| `/exit` `/quit` | 退出 |

完整列表见 `/help` 或 [使用指南](../user-guide/usage.md)。

## 状态栏

```
[⚠ BYPASS]  provider/model · ~/workdir · 1.2k tok   [⠋ 流式输出 | End 回到底部 | Ctrl+Y 复制]
```

- 左侧：provider/model、工作目录、token 用量、auto-run 提示
- `⚠ BYPASS`（红色）：`--dangerously-skip-permissions` 激活时显示
- 右侧（hint 区）：按优先级互斥显示 — 流式输出 > 未到底部 > 可复制提示

## 启动

```go
func Run(ag *agent.Agent, trustOpts trust.Options) error {
    InitTerminal()          // 检测终端能力、设置符号集与颜色方案
    m := New(ag, trustOpts)
    opts := []tea.ProgramOption{tea.WithMouseCellMotion()}
    if useAltScreen() {
        opts = append(opts, tea.WithAltScreen())
    }
    p := tea.NewProgram(&m, opts...)
    if !ag.AutoApprove() {
        ag.SetApprovalHandler(m.resolveApproval)
    }
    _, err := p.Run()
    return err
}
```

- `WithMouseCellMotion()` — 捕获滚轮事件；Shift+拖拽可原生选中文字
- `useAltScreen()` — SSH 无图形 / Windows 经典 CMD 时自动禁用备用屏幕
- `ag.AutoApprove()` — `--dangerously-skip-permissions` 时跳过安装交互审批器

## 环境变量

| 变量 | 说明 |
|------|------|
| `OPENTMD_NO_ALT_SCREEN=1` | 禁用备用屏幕（可鼠标选中历史输出） |
| `OPENTMD_ALT_SCREEN=1` | 强制启用备用屏幕 |
| `OPENTMD_ASCII=1` | 强制使用 ASCII 符号集（适用于任何平台） |
| `NO_COLOR` | 禁用颜色输出 |
