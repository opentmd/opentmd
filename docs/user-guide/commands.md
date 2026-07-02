# CLI 命令参考

## 根命令

```bash
opentmd [flags]
opentmd [command]
```

### 全局标志

| 标志 | 缩写 | 说明 |
|------|------|------|
| `--prompt` | `-p` | Prompt 模式（非交互） |

## 子命令

### `opentmd login`

配置 Provider API Key。

```bash
opentmd login                         # 交互式选择 Provider
opentmd login --provider deepseek     # 直接指定 Provider
opentmd login --oauth                 # OAuth 登录（即将上线）
```

| 标志 | 说明 |
|------|------|
| `--provider` | Provider 名称 (deepseek, openai, claude, glm, qwen, ollama) |
| `--oauth` | 使用 OpenTMD OAuth |

### `opentmd config`

显示当前配置信息。

```bash
opentmd config
```

输出示例：
```
Config file: /home/user/.opentmd/config.toml

  provider: deepseek
  model:    deepseek-chat
  ...
```

### `opentmd daemon`

启动 HTTP 服务，供 VS Code 扩展等 IDE 集成使用。

```bash
opentmd daemon --port 13456 -C ./my-project
```

| 端点 | 方法 | 说明 |
|------|------|------|
| `/health` | GET | 健康检查 |
| `/chat` | POST | SSE 聊天（`session_id`、`work_dir`） |
| `/sessions` | GET | 列出历史 session |
| `/sessions/new` | POST | 创建新 session |
| `/config` | GET | 当前 provider/model/LSP 配置摘要 |
| `/config/reload` | POST | 从磁盘重载 config.toml 并更新所有 session |
| `/lsp/status` | GET | LSP 状态（JSON，含 `text` 可读摘要） |
| `/lsp/reload` | POST | 重启所有活跃 session 的 LSP 客户端 |
| `/mcp reload` | POST | 重连 MCP 服务器 |
| `/mcp/status` | GET | MCP 服务器连接状态（JSON） |
| `/permission/respond` | POST | 响应权限审批请求 |

## 使用示例

```bash
# 交互模式
cd my-project && opentmd

# Prompt 模式
opentmd -p "解释这段代码"

# 配置 Provider
opentmd login --provider openai

# 查看配置
opentmd config

# 启动 Daemon（IDE 集成）
opentmd daemon --port 13456 -C ./my-project
curl -s http://127.0.0.1:13456/lsp/status | jq .

# Prompt 模式 + 管道
opentmd -p "检查代码风格问题" | tee report.txt
```

## 退出码

| 退出码 | 说明 |
|--------|------|
| 0 | 正常退出 |
| 1 | 一般错误 |
