# 配置指南

## 配置文件

路径: `~/.opentmd/config.toml`

### 完整配置

```toml
[default]
provider = "deepseek"
model = "deepseek-chat"

[providers.deepseek]
base_url = "https://api.deepseek.com"
api_key = "sk-..."

[providers.openai]
base_url = "https://api.openai.com/v1"
api_key = "sk-..."

[providers.claude]
base_url = "https://openrouter.ai/api/v1"
api_key = "sk-..."

[providers.glm]
base_url = "https://open.bigmodel.cn/api/paas/v4"
api_key = ""

[providers.qwen]
base_url = "https://dashscope.aliyuncs.com/compatible-mode/v1"
api_key = ""

[providers.ollama]
base_url = "http://127.0.0.1:11434"
api_key = "ollama"

[session]
persist = true

[tui]
theme = "dark"
stream = true

[agent]
max_iterations = 10

[lsp]
enabled = true
auto_detect = true
diagnostics_settle_delay_ms = 300
warmup_file_limit = 40

# 可选：按扩展名覆盖内置语言服务器
# [lsp.servers.go]
# command = "gopls"
# args = ["serve"]
```

## LSP 语言服务器

OpenTMD 内置 LSP 客户端，供 `diagnostics` 工具与写文件后自动同步使用。需在 PATH 中安装对应语言服务器（如 `gopls`、`rust-analyzer`）。

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `lsp.enabled` | 启用 LSP 诊断 | `true` |
| `lsp.auto_detect` | 自动注册内置语言服务器 | `true` |
| `lsp.diagnostics_settle_delay_ms` | 文件同步后等待诊断的毫秒数 | `300` |
| `lsp.warmup_file_limit` | 项目级诊断预热文件数上限 | `40` |
| `lsp.servers.<ext>` | 按扩展名自定义 command/args | — |

- `auto_detect = false` 且未配置 `servers` 时，不会自动启动任何 LSP
- TUI 中 `/lsp` 查看状态，`/lsp reload` 重启 LSP 客户端
- Daemon：`GET /lsp/status`（JSON）、`POST /lsp/reload`

内置映射：`.go` → gopls，`.rs` → rust-analyzer，`.ts/.js` → typescript-language-server，`.py` → pylsp，`.java` → jdtls

## 支持的 Provider

| Provider | 端点 | 需 API Key |
|----------|------|-----------|
| **DeepSeek** | `https://api.deepseek.com` | ✅ |
| **OpenAI** | `https://api.openai.com/v1` | ✅ |
| **Claude** (via OpenRouter) | `https://openrouter.ai/api/v1` | ✅ |
| **GLM (智谱)** | `https://open.bigmodel.cn/api/paas/v4` | ✅ |
| **Qwen (通义千问)** | `https://dashscope.aliyuncs.com/compatible-mode/v1` | ✅ |
| **Ollama (本地)** | `http://127.0.0.1:11434` | ❌ |

所有 Provider 使用 OpenAI 兼容 API 接口。

## 配置方式

### 方式一：首次启动向导

首次运行 `opentmd` 自动进入 3 步配置向导。

### 方式二：登录命令

```bash
opentmd login                     # 交互式选择
opentmd login --provider deepseek # 直接指定
opentmd login --oauth             # OAuth（即将上线）
```

### 方式三：手动编辑

直接编辑 `~/.opentmd/config.toml`。

## 配置项说明

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `default.provider` | 默认 Provider | deepseek |
| `default.model` | 默认模型 | deepseek-chat |
| `session.persist` | 会话持久化 | true |
| `tui.theme` | 界面主题 | dark |
| `tui.stream` | 流式输出 | true |
| `agent.max_iterations` | 最大推理轮次 | 10 |
| `lsp.enabled` | 启用 LSP 诊断 | true |
| `lsp.auto_detect` | 自动检测语言服务器 | true |
| `lsp.diagnostics_settle_delay_ms` | 诊断 settle 延迟 (ms) | 300 |
| `lsp.warmup_file_limit` | 项目预热文件数 | 40 |

## 环境变量

| 变量 | 说明 |
|------|------|
| `OPENTMD_NO_ALT_SCREEN` | 设为 `1` 禁用 TUI 备用屏幕（便于终端内选中复制） |
| `DEEPSEEK_API_KEY` 等 | 部分 Provider 支持从环境变量读取 API Key |

## 数据目录

`~/.opentmd/` 结构见 [项目结构 — 配置与数据目录](../development/project-structure.md#配置与数据目录)。
