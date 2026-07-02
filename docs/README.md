# OpenTMD-Cli 文档中心

> 开源终端 AI 编码助手 — Go 单二进制部署。

**在线文档**: [www.opentmd.com](https://www.opentmd.com) · [opentmd.com](https://opentmd.com)

本地构建站点：`cd docs/site && npm install && npm run build && npm run serve`

域名与 DNS 配置见 [docs/site/DNS.md](site/DNS.md)。

## 快速导航

| 场景 | 文档 |
|------|------|
| 安装 | [安装指南](user-guide/installation.md) |
| 卸载 | [卸载指南](user-guide/uninstall.md) |
| 日常使用 | [使用指南](user-guide/usage.md) |
| CLI 参考 | [命令参考](user-guide/commands.md) |
| 配置 | [配置指南](user-guide/configuration.md) |
| 构建发布 | [构建指南](development/build.md) · [发布指南](development/release.md) |

## 用户指南

| 文档 | 说明 |
|------|------|
| [README](../README.md) | 项目概览、快速开始 |
| [安装指南](user-guide/installation.md) | curl / npm / 源码安装 |
| [卸载指南](user-guide/uninstall.md) | 分组卸载、备份与回滚 |
| [使用指南](user-guide/usage.md) | TUI、Headless、记忆、压缩 |
| [命令参考](user-guide/commands.md) | CLI 参数与子命令 |
| [配置指南](user-guide/configuration.md) | config.toml、LSP、Provider |
| [Docker 部署](../docker/README.md) | Daemon 容器化 |

## 架构设计

| 文档 | 说明 |
|------|------|
| [架构概览](architecture/overview.md) | 模块划分与数据流 |
| [Agent 编排](architecture/agent.md) | 多步循环、工具调度 |
| [配置系统](architecture/config.md) | TOML、Provider 预设 |
| [Provider 系统](architecture/provider.md) | LLM Provider 接口 |
| [工具系统](architecture/tools.md) | 工具注册与执行 |
| [TUI 终端界面](architecture/tui.md) | Bubble Tea、斜杠命令 |
| [会话管理](architecture/session.md) | JSON 持久化 |
| [持久记忆](architecture/memory.md) | memory.md 全局 + 项目 |
| [对话压缩](architecture/compaction.md) | `/compact` 摘要策略 |
| [Hooks 系统](architecture/hooks.md) | Pre/Post ToolUse |
| [权限模型](architecture/permission.md) | 审批流 |
| [MCP 协议](architecture/mcp.md) | 外部工具集成 |
| [LSP 诊断](architecture/lsp.md) | 语言服务器客户端 |
| [Daemon 服务](architecture/daemon.md) | HTTP SSE API |
| [语义分析](architecture/semantic.md) | tree-sitter 索引 |
| [Shell 执行](architecture/shell-exec.md) | 命令引擎 |
| [文件 I/O](architecture/fileio.md) | 读写与编辑 |
| [设置向导](architecture/setup-wizard.md) | 首次启动 |

## 开发指南

| 文档 | 说明 |
|------|------|
| [项目指令](../OPENTMD.md) | 技术栈、目录结构、规范（`OPENTMD.md`） |
| [项目结构](development/project-structure.md) | 目录布局 |
| [构建指南](development/build.md) | 编译与部署 |
| [发布指南](development/release.md) | GitHub Release + npm |
| [测试指南](development/testing.md) | 测试方法 |
| [文档站点](site/README.md) | GitHub Pages 构建 |
| [npm 包](../packages/npm/README.md) | `@opentmd/cli` 发布说明 |

## 产品文档

| 文档 | 说明 |
|------|------|
| [PRD](prd/OpenTMD-Cli%20产品需求文档.md) | 产品需求文档 |
