# Daemon 服务

OpenTMD 支持以 HTTP SSE Daemon 模式运行，用于远程调用和 IDE 集成。

## 启动

```bash
opentmd daemon --port 13456 -C ./my-project
```

## API 端点

| 端点 | 方法 | 说明 |
|------|------|------|
| `/health` | GET | 健康检查 |
| `/chat` | POST | SSE 聊天（`session_id` 续聊，`work_dir` 指定项目） |
| `/sessions` | GET | 列出历史会话 |
| `/sessions/new` | POST | 创建新会话 |
| `/config` | GET | 当前配置摘要 |
| `/config/reload` | POST | 重载 config.toml |
| `/lsp/status` | GET | LSP 状态（JSON） |
| `/lsp/reload` | POST | 重启 LSP 客户端 |
| `/mcp/status` | GET | MCP 服务器状态 |
| `/mcp/reload` | POST | 重连 MCP 服务器 |
| `/permission/respond` | POST | 响应审批请求 |

## SSE 事件

| 事件 | 说明 |
|------|------|
| `session` | 会话 ID |
| `text` | 文本块 |
| `tool_start` | 工具开始执行 |
| `lsp_connect` | LSP 连接状态 |
| `permission_request` | 审批请求 |
| `done` | 完成 |
| `error` | 错误 |

## Docker 部署

```bash
# 构建
go build -o docker/opentmd ./cmd/opentmd/
docker build -t opentmd-daemon -f docker/Dockerfile-daemon docker/

# 运行
docker run -d --name opentmd-daemon \
  -p 13456:13456 \
  -v ~/.opentmd:/root/.opentmd \
  -v $(pwd):/workspace \
  opentmd-daemon

# 验证
curl http://localhost:13456/health
```

详情见 [Docker 部署文档](../docker/README.md)。

## 实现参考

参考 AtomCode `docs/atomcode/crates/atomcode-daemon/`。
