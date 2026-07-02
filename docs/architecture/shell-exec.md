# Shell 执行引擎

## 概述

Shell 执行模块位于 `internal/shell/`，提供安全的 Shell 命令执行能力，支持超时控制和输出截断。

## 核心类型

```go
type Result struct {
    Stdout   string
    Stderr   string
    ExitCode int
}
```

## 关键函数

### `Run()`

```go
func Run(ctx context.Context, command string, workDir string,
         timeout time.Duration, maxOutput int) (*Result, error)
```

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `command` | - | Shell 命令字符串 |
| `workDir` | - | 工作目录 |
| `timeout` | 60s | 执行超时 |
| `maxOutput` | 32KB | 输出截断上限 |

### `FormatResult()`

将 `Result` 格式化为可读字符串：

```
exit_code: 0
stdout:
<输出内容>
```

## 安全特性

1. **超时控制** — 通过 `context.WithTimeout` 实现，超时自动取消
2. **输出截断** — 默认 32KB 上限，防止大输出堵塞
3. **Shell 注入防护** — 使用 `exec.CommandContext`，参数与命令分离
4. **空命令检查** — 空命令返回错误

## 使用示例

```go
result, err := shell.Run(ctx, "ls -la", "/home/user/project", 30*time.Second, 64*1024)
if err != nil {
    // 处理错误（超时、命令不存在等）
}
fmt.Println(shell.FormatResult(result))
```

## 集成

由 `tool.Registry` 中的 `run_shell` 工具调用，Agent 在编排循环中使用。
