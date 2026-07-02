# OpenTMD-Cli 文档中心

> 开源终端 AI 编码助手 — Go 单二进制部署，参考 [AtomCode](atomcode/) 架构。

## 🏗️ 架构设计

| 文档 | 说明 |
|------|------|
| [架构概览](architecture/overview.md) | 整体架构与模块划分、数据流 |
| [Agent 编排](architecture/agent.md) | Agent 多步循环、工具调度、权限审批 |
| [配置系统](architecture/config.md) | TOML 配置文件、Provider 预设、LSP 配置 |
| [Provider 系统](architecture/provider.md) | LLM Provider 接口与工厂（OpenAI / Ollama） |
| [工具系统](architecture/tools.md) | 工具注册、执行、权限分类 |
| [TUI 终端界面](architecture/tui.md) | Bubble Tea 用户界面、斜杠命令、模态框 |
| [会话管理](architecture/session.md) | JSON 文件持久化、恢复、会话列表 |
| [Shell 执行](architecture/shell-exec.md) | Shell 命令引擎、超时、安全拦截 |
| [文件 I/O](architecture/fileio.md) | 文件读取（大文件截断）、写入、编辑 |
| [设置向导](architecture/setup-wizard.md) | 首次启动配置向导、OAuth 登录 |
| [Hooks 系统](architecture/hooks.md) | PreToolUse / PostToolUse / UserPromptSubmit |
| [内存记忆](architecture/memory.md) | memory.md 持久记忆（全局 + 项目） |
| [Daemon 服务](architecture/daemon.md) | HTTP SSE 服务、多会话管理 |
| [权限模型](architecture/permission.md) | AutoApprove / RequireApproval / 审批流 |
| [MCP 协议](architecture/mcp.md) | 外部 MCP 服务器集成 |
| [LSP 诊断](architecture/lsp.md) | 语言服务器协议客户端 |
| [语义分析](architecture/semantic.md) | tree-sitter 多语言代码索引 |

## 📖 用户指南

| 文档 | 说明 |
|------|------|
| [README](../../README.md) | 快速开始、CLI 参数、TUI 命令 |
| [安装脚本](../../scripts/install.sh) | 一键安装脚本（curl \| bash） |
| [Docker 部署](../../docker/README.md) | Docker 镜像构建与运行 |

## 🔧 开发指南

| 文档 | 说明 |
|------|------|
| [项目指令](.opentmd.md) | 技术栈、目录结构、构建命令、代码规范 |
| [Architecture Overview](architecture/overview.md) | 模块划分与数据流 |
| [AtomCode 参考](atomcode/README.md) | Rust 版 AtomCode 完整实现参考 |

## 📦 AtomCode 参考实现

完整 Rust 项目源码位于 `docs/atomcode/`，包含：

| 模块 | 说明 |
|------|------|
| [atomcode-core](atomcode/crates/atomcode-core/) | 核心引擎 — Agent、工具、对话、语义分析 |
| [atomcode-capabilities](atomcode/crates/atomcode-capabilities/) | L1 能力层 — Provider、MCP、Skills、Memory |
| [atomcode-cli](atomcode/crates/atomcode-cli/) | CLI 入口、自更新 |
| [atomcode-daemon](atomcode/crates/atomcode-daemon/) | HTTP SSE 服务 |
| [atomcode-review](atomcode/crates/atomcode-review/) | 代码审查引擎（50+ 语言规则） |
| [atomcode-tuix](atomcode/crates/atomcode-tuix/) | Retained-mode 终端 UI |
| [webui](atomcode/webui/) | React + TypeScript Web 界面 |
| [site](atomcode/site/) | 文档站点 |
| [docs](atomcode/docs/) | 架构、Hooks、MCP、安全、开发计划 |
| [docker](atomcode/docker/) | Docker 镜像配置 |
| [scripts](atomcode/scripts/) | 构建、安装、发布脚本 |
