# CLI 命令参考

## 根命令

```bash
opentmd [flags]
opentmd [command]
```

### 全局标志

| 标志 | 缩写 | 说明 |
|------|------|------|
| `--prompt` | `-p` | Headless 模式（非交互，输出到 stdout） |
| `--prompt-file` | | 从文件读取 prompt |
| `--dir` | `-C` | 工作目录 |
| `--continue` | `-c` | 继续最近 session |
| `--verbose` | `-v` | 工具调用详情输出到 stderr |
| `--model` | | 临时覆盖模型 |
| `--provider` | | 临时覆盖 Provider |
| `--trust` | | 信任当前工作目录（跳过确认） |
| `--version` | | 显示版本号 |
| `--help` | `-h` | 显示帮助 |

## 子命令

### `opentmd login`

配置 Provider API Key。

```bash
opentmd login
opentmd login --provider deepseek
opentmd login --oauth          # OAuth 登录（即将上线）
```

### `opentmd config`

显示当前配置摘要（provider、model、LSP、API Key 掩码等）。

```bash
opentmd config
```

### `opentmd mcp`

管理 MCP 服务器配置（`~/.opentmd/mcp.json`）。

```bash
opentmd mcp list
opentmd mcp show
opentmd mcp add myserver --command npx --args @modelcontextprotocol/server-filesystem,/path
```

### `opentmd daemon`

启动 HTTP 服务，供 VS Code 扩展等 IDE 集成使用（仅监听 loopback）。

```bash
opentmd daemon --port 13456 -C ./my-project
```

| 端点 | 方法 | 说明 |
|------|------|------|
| `/health` | GET | 健康检查 |
| `/chat` | POST | SSE 聊天（`session_id`、`work_dir`） |
| `/sessions` | GET | 列出历史 session |
| `/sessions/new` | POST | 创建新 session |
| `/config` | GET | 当前配置摘要 |
| `/config/reload` | POST | 重载 config.toml |
| `/lsp/status` | GET | LSP 状态（JSON） |
| `/lsp/reload` | POST | 重启 LSP 客户端 |
| `/mcp/status` | GET | MCP 服务器状态 |
| `/mcp/reload` | POST | 重连 MCP 服务器 |
| `/permission/respond` | POST | 响应权限审批 |

SSE 事件：`session` · `text` · `tool_start` · `lsp_connect` · `permission_request` · `done` · `error`

## TUI 斜杠命令

完整列表见 [使用指南 — TUI 命令](usage.md#tui-斜杠命令)。

常用：

| 命令 | 说明 |
|------|------|
| `/help` | 显示全部命令 |
| `/session list/new/<id>` | 会话管理 |
| `/remember` `/forget` `/memory` | 持久记忆 |
| `/compact [focus]` | LLM 摘要压缩旧对话 |
| `/mcp reload` | 重连 MCP |
| `/lsp` `/lsp reload` | LSP 状态与重启 |
| `/copy` `/export` | 复制 / 导出会话 |

## 使用示例

```bash
# 交互模式
cd my-project && opentmd

# Headless
opentmd -C ./my-project -p "运行 go test 并总结"
opentmd --prompt-file task.txt -v

# 继续上次对话
opentmd -c

# Daemon + API 探测
opentmd daemon --port 13456 &
curl -s http://127.0.0.1:13456/config | jq .
```

## 退出码

| 退出码 | 说明 |
|--------|------|
| 0 | 正常退出 |
| 1 | 一般错误 |
