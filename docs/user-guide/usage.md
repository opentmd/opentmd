# 使用指南

## 交互模式 (TUI)

```bash
cd your-project && opentmd
```

首次运行会进入 **3 步启动向导**（选择 Provider → 输入 API Key → 确认模型），配置保存在 `~/.opentmd/config.toml`。

### TUI 斜杠命令

| 命令 | 说明 |
|------|------|
| `/help` | 显示全部命令 |
| `/keys` | 显示快捷键 |
| `/status` | provider / 目录 / session / LSP 摘要 |
| `/config` | 配置文件路径与摘要 |
| `/cost` | Token 用量与估算费用 |
| `/session` | 当前 session 信息 |
| `/session list` | 列出历史 session |
| `/session new` | 新建 session |
| `/session <id>` | 切换到指定 session |
| `/resume` | 恢复历史 session |
| `/clear` | 清屏（保留对话上下文） |
| `/undo` | 撤销上一轮文件修改 |
| `/auto-run on/off/status` | 自动执行工具（跳过确认） |
| `/model` `/provider` | 切换 provider / model |
| `/login` | 配置 Provider 与 API Key |
| `/cd <dir>` | 切换工作目录 |
| `/reload` | 重载 config.toml |
| `/lsp` | LSP 语言服务器状态 |
| `/lsp reload` | 重启 LSP 客户端 |
| `/skills` | 列出 Skills |
| `/mcp` | MCP 服务器状态 |
| `/mcp reload` | 重连 MCP 服务器 |
| `/plan` `/build` | 切换只读探索 / 可执行构建模式 |
| `/copy` | 复制最近 AI 回复 |
| `/copy all` | 复制完整会话 |
| `/export [file]` | 导出会话到文件 |
| `/remember <text>` | 记忆全局事实 |
| `/remember project <text>` | 记忆项目事实 |
| `/forget <keyword>` | 删除匹配记忆 |
| `/memory` | 查看所有记忆 |
| `/compact [focus]` | LLM 摘要压缩旧对话 |
| `/exit` `/quit` | 退出 |

输入 `@` 可补全项目文件路径；输入 `/` 触发斜杠命令补全菜单。

### 快捷键

| 按键 | 功能 |
|------|------|
| `Enter` | 发送 |
| `Ctrl+J` / `Alt+Enter` | 换行 |
| `Tab` | 斜杠命令补全 |
| `Esc` | 取消输出 / 清空输入 |
| `Ctrl+O` | 切换工具详情显示 |
| `Ctrl+L` | 清屏（保留对话） |
| `Ctrl+Y` | 复制最近 AI 回复到剪贴板 |
| `Ctrl+G` / `End` | 滚到最新 |
| `Ctrl+C` | 取消；连按两次退出 |

TUI 默认使用备用屏。**复制输出**有以下方式：

- `Ctrl+Y` — 复制最近一条 AI 回复；`/copy all` 复制完整会话
- `Shift+拖拽` — 在大多数终端（xterm、GNOME Terminal、iTerm2、Windows Terminal）中绕过鼠标捕获，直接原生选中
- 内容同时写入 `~/.opentmd/last-copy.txt`，无剪贴板环境也可得到

### 权限审批

对 bash / 写文件 / MCP 等敏感操作：

| 按键 | 含义 |
|------|------|
| `y` | 允许一次 |
| `a` | 本会话允许 |
| `n` | 拒绝 |

**自动批准模式（跳过所有权限提示）**

```bash
opentmd -y                        # TUI 交互模式，所有工具自动批准
opentmd -y -p "运行全量测试"          # Headless，无任何提示
```

启用 `-y` 后：
- 所有工具调用（包括高危操作）自动允许，无需手动确认
- TUI 状态栏显示红色 `⚠ BYPASS` 徽章
- 隐含 `--trust`，自动信任当前工作目录

> ⚠️ **警告**：该模式不拦截任何操作，仅在您信任流水线和代理内置安全约束的环境中使用。

## Headless 模式（CI / 脚本）

```bash
opentmd -C ./my-project -p "运行 go test 并总结"
opentmd --prompt-file task.txt -v
opentmd -c -p "继续上次的任务"
```

直接输出到 stdout，`-v` 将工具调用详情写入 stderr。

## Agent 工具

Agent 自动循环调用工具直至任务完成。完整列表见 [README — Agent 工具](../../README.md#agent-工具对齐-atomcode)。

典型工作流：`read/grep/glob` → `edit/write` → `bash`（测试）→ 验证 → 回复。

## 持久记忆

```bash
/remember 偏好中文回复
/remember project 使用 go test ./...
/memory
/forget 中文
```

- 全局：`~/.opentmd/memory.md`
- 项目：`<root>/.opentmd/memory.md`

详见 [架构 — 持久记忆](../architecture/memory.md)。

## 对话压缩

长会话中上下文过大时：

```bash
/compact              # 摘要合并旧消息
/compact auth 模块    # 聚焦某主题的摘要
```

保留最近 6 条消息，旧消息合并为 `[Previous conversation summary]` 注入会话。详见 [架构 — 对话压缩](../architecture/compaction.md)。

## 配置查看

```bash
opentmd config
```
