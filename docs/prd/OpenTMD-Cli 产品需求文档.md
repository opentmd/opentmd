# OpenTMD-Cli 产品需求文档（PRD）

## 一、项目概述

### 1.1 项目名称

**OpenTMD-Cli**



### 1.2 项目定位

OpenTMD-Cli 是一个运行在 Linux 终端里的 AI 编码助手。

定位类似 [Claude Code](https://www.anthropic.com/claude-code)：

* 用户进入终端
* 用自然语言描述任务
* AI 自动理解需求
* 自主读取项目代码
* 编辑文件
* 执行命令
* 自我验证结果

目标是打造：

**开源 + 本地终端 + AI Coding Agent + Go 单二进制 + 可扩展插件生态**



### 1.3 技术目标

要求：

* 使用 Go 开发
* Linux 优先（Ubuntu 22）
* 编译为单二进制
* 默认支持 DeepSeek
* 支持扩展 OpenAI-compatible provider
* TUI 风格参考 Claude Code
* 100% AI 生成代码优先



## 二、核心目标

### MVP 第一阶段

完成一个可用 CLI Agent：

支持：

* 自然语言提问
* 调用模型
* 流式输出
* 多轮对话
* 读取本地文件
* Shell 命令执行
* 基础 slash command



### 第二阶段

增强体验：

* syntax highlight
* session persistence
* 文件 diff
* MCP
* plugin



### 第三阶段

生态：

* skill marketplace
* Claude Code plugin compatibility
* issue / workflow automation



# 三、技术架构

推荐结构：

```txt
opentmd-cli/

cmd/
    opentmd/

internal/

    tui/
    agent/
    model/
    provider/
    session/
    tool/
    shell/
    file/
    config/
    command/
    plugin/
    mcp/

pkg/

assets/

```



## 模块说明



### 1）cmd/opentmd

入口：

```bash
opentmd
opentmd -p "介绍仓库"
opentmd login
opentmd config
```

职责：

* CLI 参数解析
* 初始化 config
* 启动 TUI
* 执行 prompt mode

建议：

Go：

```go
cobra
```



### 2）config

路径：

```bash
~/.opentmd/config.toml
```

示例：

```toml
[default]
provider = "deepseek"
model = "deepseek-chat"

[providers.deepseek]
base_url = ""
api_key = ""

[providers.openai]
base_url = ""
api_key = ""

[session]
persist = true

[tui]
theme = "dark"
stream = true
```

职责：

* 读取
* 初始化
* 热更新



### 3）provider

统一模型接口：

```go
type Provider interface {
    Chat(ctx context.Context, req ChatRequest) (<-chan StreamChunk, error)
}
```

实现：

* DeepSeek
* OpenAI compatible
* Ollama
* Claude-compatible



### 4）agent

核心编排器

负责：

* 用户输入
* 拼装上下文
* 调用模型
* 决策 tool
* 执行 tool
* 汇总回复

流程：

```txt
用户输入
↓

session context

↓

provider.Chat()

↓

检测 tool 调用

↓

执行 tool

↓

继续推理

↓

输出
```



### 5）tui

终端 UI

推荐：

[Bubble Tea](https://github.com/charmbracelet/bubbletea?utm_source=chatgpt.com)

配套：

* lipgloss
* glamour

界面：

```txt
┌────────────────────┐
│ OpenTMD            │
├────────────────────┤
│ user >             │
│ assistant...       │
│                    │
├────────────────────┤
│ input:             │
└────────────────────┘
```



# 四、功能需求



## 4.1 流式输出

优先级：

P0

效果：

逐字打印：

```txt
正在分析...
读取 package.json...
发现 Vue3...
建议如下...
```

要求：

* token chunk 实时渲染
* 自动换行
* 打字体验自然



## 4.2 多轮会话记忆

优先级：

P0

要求：

当前 session 保留：

```txt
用户 → 分析项目
AI → 回复
用户 → 修改 utils
AI 继续
```

存储：

```bash
~/.opentmd/sessions/
```

格式：

json

支持：

恢复：

```bash
/session
```



## 4.3 本地文件读取

优先级：

P0

支持：

```txt
README.md
package.json
src/*
```

功能：

* 单文件
* 多文件
* glob
* 目录树

限制：

大文件截断



## 4.4 文件编辑

优先级：

P1

支持：

* overwrite
* append
* patch

展示：

diff：

```diff
- old
+ new
```

确认：

```txt
Apply changes? y/n
```



## 4.5 Shell 执行

优先级：

P1

支持：

```bash
npm install
go test
pytest
```

展示 stdout

限制：

超时



## 4.6 代码高亮

优先级：

P1

支持：

* Go
* Python
* JS
* TS
* Vue
* JSON



## 4.7 Prompt Mode

优先级：

P0

命令：

```bash
opentmd -p "介绍仓库"
```

输出 stdout

适合：

* CI
* shell



## 4.8 Slash Commands

优先级：

P1

支持：

```bash
/help
/model
/session
/undo
/clear
/config
/issue
/exit
```



## 4.9 Undo

优先级：

P2

撤销：

* 上次文件改动



## 4.10 图片支持

优先级：

P2

支持：

* Ctrl+V
* 文件路径

传模型：

vision



# 五、MCP

优先级：

P2

配置：

```json
{
  "servers": []
}
```

支持：

* GitHub
* DB
* Playwright

流程：

读取配置

连接 server

暴露 tool



# 六、Plugin

优先级：

P2

目录：

```bash
~/.opentmd/plugins
```

支持：

* git install
* skill
* hooks



# 七、错误处理

要求：

* API 错误
* 网络错误
* 超时
* Ctrl+C

提示友好



# 八、性能要求

启动：

< 200ms

流式延迟：

< 1s

内存：

< 100MB



# 九、发布

平台：

Linux

构建：

```bash
go build
```

产物：

```bash
opentmd
```

安装：

```bash
curl ... | bash
```



# 十、开发顺序

## Sprint1

完成：

* CLI
* config
* provider
* stream output
* session
* prompt mode



## Sprint2

完成：

* file read
* shell
* slash command
* syntax highlight



## Sprint3

完成：

* patch
* undo
* plugin
* mcp



# 十一、验收标准

满足：

## 基础

```bash
opentmd
```

进入 UI



## 提问

```txt
分析这个仓库
```

返回结果



## 流式

逐字输出



## session

上下文记忆



## 文件

可读



## prompt

```bash
opentmd -p
```

可执行



## binary

单文件运行



# 最终目标

打造：

**像 Claude Code 一样顺手**
+
**比 Claude Code 更开放**
+
**更适合中国开发者生态**
+
**完全开源可扩展**
