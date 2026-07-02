# 项目结构

## 目录布局

```
opentmd-cli/
├── cmd/opentmd/              # CLI 入口（Cobra）
│   ├── main.go               # 根命令、login、config
│   ├── daemon.go             # daemon 子命令
│   ├── mcp.go                # mcp 子命令
│   └── output.go             # Headless 输出
├── internal/                 # 内部包
│   ├── agent/                # Agent 编排、compact
│   ├── compaction/           # 对话压缩
│   ├── config/               # TOML 配置
│   ├── daemon/               # HTTP SSE 服务
│   ├── deepseek/             # OpenAI 兼容客户端
│   ├── hook/                 # Hooks 系统
│   ├── llm/                  # LLM 类型与接口
│   ├── lsp/                  # LSP 客户端
│   ├── mcp/                  # MCP 客户端
│   ├── memory/               # 持久记忆
│   ├── permission/           # 权限审批
│   ├── provider/             # Provider 工厂
│   ├── semantic/             # tree-sitter 符号分析
│   ├── session/              # 会话持久化
│   ├── setup/                # 首次启动向导
│   ├── skill/                # Skills 系统
│   ├── tool/                 # 工具注册与执行
│   ├── tui/                  # Bubble Tea TUI
│   └── web/                  # web_search / web_fetch
├── packages/npm/             # npm 分发（@opentmd/cli）
├── extensions/vscode-opentmd/ # VS Code 扩展
├── docker/                   # Docker 镜像
├── scripts/                  # install / build / deploy / uninstall
├── docs/                     # 文档中心
│   ├── user-guide/           # 用户指南
│   ├── architecture/         # 架构设计
│   ├── development/          # 开发指南
│   └── atomcode/             # AtomCode 参考实现
├── opentmd                   # 仓库内快捷启动脚本
├── go.mod
└── LICENSE
```

## 依赖关系

```
cmd/opentmd
  └── internal/agent
        ├── internal/tool ── internal/{fileio,shell,grep,glob,lsp,mcp,skill,...}
        ├── internal/session
        ├── internal/memory
        ├── internal/compaction
        ├── internal/hook
        └── internal/provider ── internal/deepseek
```

## 脚本说明

| 脚本 | 用途 |
|------|------|
| `scripts/install.sh` | curl / 本地安装 |
| `scripts/uninstall.sh` | 分组卸载 |
| `scripts/build.sh` | 编译与交叉发布 |
| `scripts/deploy-local.sh` | 本地部署到 PATH |
| `scripts/run-local.sh` | 仓库内 `opentmd` 启动 |
| `packages/npm/scripts/build.sh` | npm 平台包发布 |

## 配置与数据目录

`~/.opentmd/` 由 `config.Ensure()` 自动创建：

```
~/.opentmd/
├── config.toml       # 主配置
├── mcp.json          # MCP 服务器
├── hooks.json        # Hooks
├── memory.md         # 全局记忆
├── sessions/         # 会话 JSON
├── skills/           # Skills 模板
└── plugins/          # 插件目录
```
