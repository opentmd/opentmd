# 架构概览

## 整体架构

OpenTMD-Cli 采用**模块化单体架构**，所有功能编译成单个 Go 二进制文件。

```
┌───────────────────────────────────────────────────────────┐
│                    cmd/opentmd                             │
│                   (Cobra CLI 入口)                          │
├───────────────────────────────────────────────────────────┤
│                    agent (编排器)                           │
│  ┌──────────┐ ┌──────────┐ ┌──────────────┐              │
│  │ provider  │ │ session  │ │   tool(s)    │              │
│  │ (LLM调用)  │ │ (会话管理) │ │ (read/write/  │              │
│  │           │ │          │ │  grep/bash…)  │              │
│  └──────────┘ └──────────┘ └──────────────┘              │
│  ┌──────────┐ ┌──────────┐ ┌──────────────┐              │
│  │ memory   │ │  hook    │ │  skill       │              │
│  │ (持久记忆) │ │ (钩子系统) │ │ (技能模板)    │              │
│  └──────────┘ └──────────┘ └──────────────┘              │
│  ┌──────────┐                                             │
│  │compaction│  /compact 对话压缩                           │
│  └──────────┘                                             │
├───────────────────────────────────────────────────────────┤
│     tui (Bubble Tea)   │   daemon (HTTP SSE)             │
│     setup (设置向导)    │   mcp (外部工具)                 │
│     config (TOML配置)   │   lsp (语言服务器)               │
├───────────────────────────────────────────────────────────┤
│               Go 标准库 + 第三方依赖                        │
└───────────────────────────────────────────────────────────┘
```

## 数据流

### 交互模式 (TUI)

```
用户输入 → TUI 接收
         → Agent.Chat()
           → Hook: UserPromptSubmit（可修改/阻止）
           → 加载 Session 历史
           → 构建 System Prompt（含 memory + 项目指令 + skills）
           → Provider.Chat() (流式)
           → 检测 Tool 调用
             → Hook: PreToolUse（可阻止/修改）
             → 权限检查（AutoApprove / RequireApproval）
             → 执行 Tool
             → Hook: PostToolUse（fire-and-forget）
           → 继续推理/输出
         → 流式渲染到 TUI
         → 保存到 Session
```

### Prompt 模式 (非交互)

```
opentmd -p "prompt"
         → Agent.ChatToWriter()
           → 同上流程
         → 输出到 stdout
```

### Daemon 模式 (HTTP SSE)

```
POST /chat
  → 创建/恢复 Agent
     → Agent.Chat() (同上)
       → SSE 事件流: text, tool_start, permission_request, done
  → 关闭 Agent
```

## 模块职责

| 模块 | 职责 |
|------|------|
| `cmd/opentmd` | Cobra CLI 入口，命令注册（login/config/daemon/mcp） |
| `internal/config` | TOML 配置读写，Provider 预设，LSP 配置 |
| `internal/llm` | LLM 类型定义与 Provider 接口 |
| `internal/provider` | Provider 工厂（OpenAI 兼容 / Ollama） |
| `internal/deepseek` | OpenAI 兼容 HTTP 客户端（流式/非流式） |
| `internal/agent` | Agent 多步循环编排，prompt 构建，记忆注入 |
| `internal/tool` | 工具注册表、执行运行时、权限钩子 |
| `internal/fileio` | 文件读取（大文件截断 skeleton） |
| `internal/shell` | Shell 命令执行（超时、nohup、安全拦截） |
| `internal/session` | 会话持久化（JSON 文件） |
| `internal/setup` | 首次启动配置向导 |
| `internal/tui` | Bubble Tea TUI（聊天、斜杠命令、模态框） |
| `internal/memory` | 持久记忆（memory.md 全局 + 项目） |
| `internal/hook` | Hooks 系统（PreToolUse/PostToolUse/UserPromptSubmit） |
| `internal/skill` | Skills 系统（SKILL.md 模板加载） |
| `internal/daemon` | HTTP SSE Daemon 服务 |
| `internal/mcp` | MCP 客户端（stdio 传输） |
| `internal/lsp` | LSP 客户端（gopls/rust-analyzer/tsserver/pylsp） |
| `internal/diagnostics` | LSP 实时诊断收集 |
| `internal/semantic` | tree-sitter 多语言符号分析 |
| `internal/graph` | Go 代码图谱索引（trace_callers/trace_callees） |
| `internal/grep` | 正则搜索代码 |
| `internal/glob` | 文件模式匹配 |
| `internal/permission` | 权限审批（AutoApprove/RequireApproval） |
| `internal/trust` | 工作目录信任管理 |
| `internal/pricing` | Token 费用估算 |
| `internal/web` | 网络工具（web_search / web_fetch） |
| `internal/filehistory` | 文件修改历史（/undo 恢复） |
| `internal/pathutil` | 路径解析工具 |
| `internal/project` | 项目指令加载，技术栈检测 |
| `internal/cliname` | CLI 名称工具 |
| `extensions/vscode-opentmd` | VS Code 扩展 |
