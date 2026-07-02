# OpenTMD-Cli

AI 编程助手
在你的终端中运行 Claude Code / Cursor Agent 的开源平替，住在你的终端里。连接任意 OpenAI 兼容大模型，自主读文件、改代码、跑测试、自我验证 —— 用 Go 构建。

## 快速开始

### 安装

```bash
./scripts/install.sh
# 或
curl -fsSL https://raw.githubusercontent.com/opentmd/opentmd-cli/master/scripts/install.sh | bash
```

### 配置与运行

```bash
opentmd login
cd your-project && opentmd
```

### Headless（CI / 脚本）

```bash
opentmd -C ./my-project -p "运行 go test 并总结"
opentmd --prompt-file task.txt -v
opentmd -c   # 继续最近 session
```

## Agent 工具（对齐 AtomCode）

| 工具 | 说明 |
|------|------|
| `read_file` | 读取文件（大文件截断） |
| `write_file` | 创建/覆盖文件 |
| `edit_file` | 单文件精确替换 |
| `search_replace` | 多文件批量替换 |
| `grep` | 正则搜索代码 |
| `glob` | 按模式查找文件 |
| `list_directory` | 列出目录 |
| `bash` | 执行 Shell（测试/构建） |
| `use_skill` | 加载 SKILL.md 模板 |
| `todo` | 多步任务进度跟踪 |
| `web_search` | DuckDuckGo 网页搜索 |
| `web_fetch` | 抓取公开 URL（含 SSRF 防护） |
| `list_symbols` | 多语言符号列表（Go/Rust/Python/JS 等） |
| `read_symbol` | 按符号名读取源码片段 |
| `diagnostics` | LSP 实时诊断（gopls / rust-analyzer / tsserver / pylsp） |
| `trace_callers` | Go 代码图谱：查找调用者 |
| `trace_callees` | Go 代码图谱：查找被调用者 |
| `file_dependencies` | Go 代码图谱：文件符号 |
| `mcp__<server>__<tool>` | MCP 外部工具（需审批） |

## Phase 2/3 能力

| 模块 | 说明 |
|------|------|
| **MCP** | `~/.opentmd/mcp.json` + 项目 `.mcp.json`；`opentmd mcp list/add/show` |
| **Skills** | `~/.opentmd/skills/*/SKILL.md`；TUI `/skills` |
| **Hooks** | `~/.opentmd/hooks.json`；PreToolUse / PostToolUse / UserPromptSubmit |
| **Memory** | `~/.opentmd/memory.md` + 项目 `.opentmd/memory.md`；`/remember` `/forget` `/memory` |
| **权限审批** | TUI 中 `[y/a/n]` 审批 bash/写文件/MCP |
| **代码图谱** | Go AST 索引，`trace_callers` / `file_dependencies` |
| **Daemon** | `opentmd daemon --port 13456`；HTTP/SSE `/chat` |
| **VS Code** | `extensions/vscode-opentmd/` 侧边栏 Chat + LSP 状态树 |

## 项目指令（对齐 AtomCode）

自动注入 system prompt：

- 全局：`~/.opentmd/OPENTMD.md`
- 项目：`.opentmd.md` / `OPENTMD.md` / `CLAUDE.md` / `.atomcode.md`
- 用户：`.opentmd.user.md`

## TUI 命令

| 命令 | 说明 |
|------|------|
| `/cd <dir>` | 切换工作目录 |
| `/reload` | 热加载 config.toml |
| `/status` | provider / 目录 / session / LSP 摘要 |
| `/config` | 配置文件路径与摘要 |
| `/lsp` | LSP 语言服务器状态 |
| `/lsp reload` | 重启 LSP 客户端 |
| `/model` `/provider` | 打开 provider/model 选择器 |
| `/session` | 打开 session 选择器 |
| `/session list/new/<id>` | 会话管理 |
| `/skills` | 列出 Skills |
| `/plan` `/build` | 切换只读探索 / 可执行构建模式 |
| `/mcp reload` | 重连 MCP 服务器 |
| `/copy` `/copy all` | 复制 AI 回复 / 完整会话到剪贴板 |
| `/export [file]` | 导出会话到文本文件 |
| `/remember <text>` | 记忆全局事实（`project` 前缀为项目级） |
| `/forget <keyword>` | 删除匹配记忆 |
| `/memory` | 查看所有记忆 |
| `/compact [focus]` | LLM 摘要压缩旧对话 |
| `/help` `/clear` `/exit` | 基础命令 |

TUI 使用备用屏幕，无法用鼠标选中复制；`Ctrl+Y` 复制最近 AI 回复。若需终端内选中，可设 `OPENTMD_NO_ALT_SCREEN=1`。

权限审批（bash / 写文件 / MCP）：`y` 允许一次 · `a` 本会话允许 · `n` 拒绝

## CLI 参数

| 参数 | 说明 |
|------|------|
| `-C, --dir` | 工作目录 |
| `-c, --continue` | 继续最近 session |
| `-p, --prompt` | Headless 模式 |
| `--prompt-file` | 从文件读取 prompt |
| `-v, --verbose` | 工具调用输出到 stderr |
| `--model` `--provider` | 临时覆盖配置 |

### Daemon & IDE

```bash
opentmd daemon --port 13456 -C ./my-project   # 启动 HTTP 服务（会话复用）
opentmd mcp add myserver --command npx --args @modelcontextprotocol/server-filesystem,/path
cd extensions/vscode-opentmd && npm install && npm run compile
```

Daemon API（loopback only）：

| 端点 | 说明 |
|------|------|
| `POST /chat` | SSE 聊天，`session_id` 续聊，`work_dir` 指定项目目录 |
| `POST /sessions/new` | 创建新会话 |
| `GET /sessions` | 列出历史会话 |
| `GET /config` | 当前配置摘要 |
| `POST /config/reload` | 从磁盘重载 config.toml |
| `GET /lsp/status` | LSP 状态（JSON） |
| `POST /lsp/reload` | 重启 LSP 客户端 |
| `GET /mcp/status` | MCP 服务器状态（JSON） |
| `POST /mcp/reload` | 重连 MCP 服务器 |
| `POST /permission/respond` | 响应 SSE `permission_request` 事件 |

SSE 事件：`session` · `text` · `tool_start` · `lsp_connect` · `permission_request` · `done` · `error`

## 开发

```bash
./scripts/build.sh
go test ./...
```

## License

See [LICENSE](LICENSE).
