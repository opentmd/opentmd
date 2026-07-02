# TUI 终端界面

## 概述

TUI 位于 `internal/tui/`，基于 [Bubble Tea](https://github.com/charmbracelet/bubbletea) 框架实现，提供终端交互界面。

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
┌─────────────────────────────────────────────┐
│  OpenTMD — 终端 AI 编码助手                   │
│                                              │
│  ┌─────────────────────────────────────────┐ │
│  │                                         │ │
│  │  user > 你好                           │ │
│  │  assistant > 你好！我能帮你做什么？      │ │
│  │                                         │ │
│  │  [消息列表 / Viewport]                  │ │
│  │                                         │ │
│  └─────────────────────────────────────────┘ │
│  ┌─────────────────────────────────────────┐ │
│  │ > 输入框...                             │ │
│  └─────────────────────────────────────────┘ │
│  [状态栏]                                    │
└─────────────────────────────────────────────┘
```

## 核心组件

### Model (tui.go)

Bubble Tea Model，管理：

| 字段 | 说明 |
|------|------|
| `agent` | Agent 实例引用 |
| `viewport` | 消息视口 (滚动) |
| `textarea` | 输入框 |
| `messages` | 显示消息列表 |
| `spinner` | 等待动画 |
| `err` | 错误信息 |

### 消息类型

TUI 内部通过 `tea.Msg` 传递异步事件：

| 消息 | 说明 |
|------|------|
| `chunkMsg` | 流式输出 chunk |
| `statusMsg` | 状态更新 |
| `doneMsg` | 流式输出完成 |

### 消息渲染 (wrap.go)

- `wrapPlainText()` — 智能换行，支持中日韩字符
- `msgPrefixStyle()` — 按角色设置不同样式

### 消息前缀样式

| 角色 | 前缀 | 颜色 |
|------|------|------|
| user | `user > ` | 用户样式 |
| assistant | `assistant > ` | 助手样式 |
| tool | (无前缀) | 工具样式 |
| system | (无前缀) | 系统样式 |

## Slash 命令

| 命令 | 说明 |
|------|------|
| `/help` | 显示帮助 |
| `/clear` | 清空当前会话 |
| `/session` | 显示当前 session |
| `/session list` | 列出历史 session |
| `/session new` | 新建 session |
| `/session <id>` | 切换到指定 session |
| `/copy` | 复制最近 AI 回复到剪贴板 |
| `/copy all` | 复制完整会话 |
| `/export [file]` | 导出会话到文件（默认写到工作目录） |
| `/exit` | 退出 |

快捷键：`Ctrl+C` 退出，`Ctrl+Y` 复制最近 AI 回复，`Esc` 清空输入框。

TUI 默认使用备用屏幕（`WithAltScreen`），终端内无法用鼠标选中复制。可用 `/copy`、`/export`，或设置 `OPENTMD_NO_ALT_SCREEN=1` 关闭备用屏幕以使用终端滚动与选中。

## 启动

```go
func Run(ag *agent.Agent) error {
    m := New(ag)
    p := tea.NewProgram(&m, tea.WithAltScreen(), tea.WithMouseAllMotion())
    m.program = p
    _, err := p.Run()
    return err
}
```

- `WithAltScreen()` — 使用备用屏幕（退出后恢复终端）
- `WithMouseAllMotion()` — 支持鼠标滚轮滚动
