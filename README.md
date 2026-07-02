# OpenTMD-Cli

开源终端 AI 编码助手，Go 单二进制部署。

## 快速开始

### 安装

```bash
# curl 一键安装（推荐）
curl -fsSL https://raw.githubusercontent.com/opentmd/opentmd-cli/master/scripts/install.sh | bash

# npm 全局安装
npm install -g @opentmd/cli

# 本地源码
./scripts/install.sh
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

### 卸载

```bash
curl -fsSL https://raw.githubusercontent.com/opentmd/opentmd-cli/master/scripts/uninstall.sh | bash
npm uninstall -g @opentmd/cli   # npm 用户
```

## Agent 工具

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
| `list_symbols` | 多语言符号列表 |
| `read_symbol` | 按符号名读取源码片段 |
| `diagnostics` | LSP 实时诊断 |
| `trace_callers` / `trace_callees` / `file_dependencies` | Go 代码图谱 |
| `mcp__<server>__<tool>` | MCP 外部工具（需审批） |

## 核心能力

| 模块 | 说明 |
|------|------|
| **MCP** | `~/.opentmd/mcp.json`；`opentmd mcp list/add/show` |
| **Skills** | `~/.opentmd/skills/*/SKILL.md`；`/skills` |
| **Hooks** | `~/.opentmd/hooks.json` |
| **Memory** | `memory.md` 持久记忆；`/remember` `/forget` `/memory` |
| **Compaction** | `/compact` LLM 摘要压缩旧对话 |
| **权限审批** | TUI 中 `y` / `a` / `n` |
| **Daemon** | `opentmd daemon`；HTTP/SSE API |
| **VS Code** | `extensions/vscode-opentmd/` |

## TUI 命令

| 命令 | 说明 |
|------|------|
| `/session list/new/<id>` | 会话管理 |
| `/remember` `/forget` `/memory` | 持久记忆 |
| `/compact [focus]` | 压缩对话历史 |
| `/mcp reload` | 重连 MCP |
| `/lsp` `/lsp reload` | LSP 状态 |
| `/help` | 完整命令列表 |

完整列表见 [使用指南](docs/user-guide/usage.md)。

## CLI 参数

| 参数 | 说明 |
|------|------|
| `-C, --dir` | 工作目录 |
| `-c, --continue` | 继续最近 session |
| `-p, --prompt` | Headless 模式 |
| `--prompt-file` | 从文件读取 prompt |
| `-v, --verbose` | 工具调用输出到 stderr |
| `--model` `--provider` | 临时覆盖配置 |
| `--version` | 显示版本 |

## 文档

| 文档 | 说明 |
|------|------|
| [在线文档](https://www.opentmd.com) | GitHub Pages 站点（www.opentmd.com） |
| [文档中心](docs/README.md) | 全部文档索引 |
| [安装指南](docs/user-guide/installation.md) | curl / npm / 源码 |
| [架构设计](docs/architecture/overview.md) | 模块与数据流 |
| [DNS 配置](docs/site/DNS.md) | 自定义域名设置 |

## 开发

```bash
./scripts/build.sh
go test ./...
```

## License

See [LICENSE](LICENSE).
