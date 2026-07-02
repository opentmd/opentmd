# MCP 协议集成

MCP (Model Context Protocol) 允许 OpenTMD 通过 stdio 连接外部工具服务器。

## 配置

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/path/to/project"],
      "disabled": false
    }
  }
}
```

加载路径：
- `~/.opentmd/mcp.json` — 全局
- `<project>/.opentmd/mcp.json` / `.mcp.json` — 项目级

## 工具命名

MCP 工具以 `mcp__<server>__<tool>` 格式注册：

```
mcp__filesystem__read_file
mcp__filesystem__write_file
mcp__github__create_issue
```

## CLI

```bash
opentmd mcp add myserver --command npx --args @modelcontextprotocol/server-filesystem,/path
opentmd mcp list
opentmd mcp show myserver
```

## 实现

- `internal/mcp/mcp.go` — MCP 连接管理
- `internal/mcp/client_stdio.go` — stdio 传输客户端
- `internal/mcp/status.go` — 状态报告
- `internal/tool/runtime.go:registerExtended()` — MCP 工具注册
