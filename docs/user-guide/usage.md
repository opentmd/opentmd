# 使用指南

## 交互模式 (TUI)

```bash
cd your-project && opentmd
```

进入 TUI 交互界面：

```
┌─────────────────────────────────────────────┐
│  OpenTMD — 终端 AI 编码助手                   │
│                                              │
│  user > 分析这个项目的结构                      │
│  assistant > 正在分析...                      │
│                                              │
│  > _                                         │
│                                              │
└─────────────────────────────────────────────┘
```

### TUI 命令

| 命令 | 说明 |
|------|------|
| `/help` | 显示帮助 |
| `/clear` | 清空当前会话 |
| `/session` | 显示当前 session |
| `/session list` | 列出历史 session |
| `/session new` | 新建 session |
| `/session <id>` | 切换到指定 session |
| `/exit` | 退出 |

### 快捷键

| 按键 | 功能 |
|------|------|
| `Ctrl+C` | 退出程序 |
| `Esc` | 清空输入框 |
| 鼠标滚轮 | 滚动消息历史 |

## Prompt 模式 (非交互)

```bash
opentmd -p "分析这个仓库"
opentmd --prompt "解释这段代码的作用"
```

直接输出结果到 stdout，适用于管道和脚本：

```bash
opentmd -p "解释这个项目" | grep -i "feature"
```

## Agent 能力

Agent 会自动循环调用工具直至任务完成：

| 工具 | 说明 |
|------|------|
| `read_file` | 读取本地文件（大文件自动截断） |
| `list_dir` | 列出目录内容 |
| `run_shell` | 执行 Shell 命令（带超时） |

### 使用示例

```
你 > 帮我看看这个项目的入口文件在哪里
助手 > 让我先查看一下项目结构...
[调用 list_dir 工具]
项目根目录包含 cmd/、internal/ 等目录。

你 > 读取 main.go
助手 > [调用 read_file 工具]
主入口在 cmd/opentmd/main.go，使用 Cobra 框架...

你 > 帮我编译
助手 > [调用 run_shell 工具: go build -o bin/opentmd ./cmd/opentmd/]
编译成功！
```

## 配置查看

```bash
opentmd config
```

显示当前配置信息：
```
Config file: /home/user/.opentmd/config.toml

  provider: deepseek
  model:    deepseek-chat
  theme:    dark
  stream:   true
  persist:  true
  setup:    completed=true

  [deepseek] (active)
    base_url: https://api.deepseek.com
    api_key:  set (sk-f0fc...02fc)

  Available providers:
    deepseek   DeepSeek [active]
    openai     OpenAI [configured]
    ...
```
