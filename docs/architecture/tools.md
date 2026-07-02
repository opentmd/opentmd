# 工具系统

## 概述

工具系统位于 `internal/tool/`，实现 Agent 的工具注册与执行机制，支持 Function Calling。

## 核心结构

```go
type Handler func(ctx context.Context, args json.RawMessage) (string, error)

type Registry struct {
    workDir string
    tools   map[string]toolEntry
}

type toolEntry struct {
    def    llm.ToolDefinition
    handle Handler
}
```

## 注册机制

`NewRegistry(workDir)` 初始化并注册默认工具：

```go
r.Register(llm.ToolDefinition{...}, handler)
```

每个工具包含：
- `ToolDefinition` — 工具名称、描述、参数 JSON Schema
- `Handler` — 实际执行函数

## 内置工具

### 1. read_file — 读取本地文件

```json
{
  "name": "read_file",
  "description": "Read the contents of a local file. Large files are truncated.",
  "parameters": {
    "type": "object",
    "properties": {
      "path": {
        "type": "string",
        "description": "File path relative to the project root or absolute path"
      }
    },
    "required": ["path"]
  }
}
```

### 2. run_shell — 执行 Shell 命令

```json
{
  "name": "run_shell",
  "description": "Execute a shell command in the project directory and return stdout/stderr.",
  "parameters": {
    "type": "object",
    "properties": {
      "command": {
        "type": "string",
        "description": "The shell command to execute"
      }
    },
    "required": ["command"]
  }
}
```

### 3. list_dir — 列出目录内容

```json
{
  "name": "list_dir",
  "description": "List files and directories in a given path.",
  "parameters": {
    "type": "object",
    "properties": {
      "path": {
        "type": "string",
        "description": "Directory path to list"
      }
    },
    "required": []
  }
}
```

## 工具执行流程

```
Agent 检测到 Tool Call
    ↓
解析 ToolCall.Name + Arguments (JSON)
    ↓
Registry.Execute(name, args)
    ↓
查找注册的 Handler
    ↓
执行 Handler
    ↓
返回结果字符串
```

## 集成 LLM

工具定义通过 `ChatRequest.Tools` 发送给 LLM，模型返回 `tool_calls` 指明要调用的工具和参数。

## 扩展性

可通过 `r.Register()` 注册自定义工具，只需提供 ToolDefinition + Handler 函数。
