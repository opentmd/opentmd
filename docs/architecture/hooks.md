# Hooks 系统

参考 AtomCode 的 `hooks.md` 实现。支持 **JSON CC 兼容格式**的 hooks 配置。

## 触发事件

| 事件 | 说明 |
|------|------|
| `PreToolUse` | 工具执行前 — 可阻止或修改参数 |
| `PostToolUse` | 工具执行后 — fire-and-forget |
| `UserPromptSubmit` | 用户消息提交 — 可修改或阻止 |
| `SessionStart` | 会话启动 |

## 配置

```json
{
  "hooks": {
    "my-hook": {
      "event": "pre_tool_use",
      "matcher": "write*",
      "command": "echo '{\"action\": \"allow\"}'",
      "timeout_ms": 10000,
      "disabled": false
    }
  }
}
```

加载路径：
- `~/.opentmd/hooks.json` — 全局
- `<project>/.hooks.json` / `<project>/.opentmd/hooks.json` — 项目级

## 环境变量

Hook 脚本通过环境变量接收上下文：

| 变量 | 说明 |
|------|------|
| `OPENTMD_HOOK_EVENT` | 事件类型 |
| `OPENTMD_TOOL_NAME` | 工具名 |
| `OPENTMD_HOOK_CONTEXT` | 上下文 JSON |

## 输出格式

Hook stdout 需输出 JSON：

```json
{"action": "allow"}           // 允许（pre_tool_use 默认）
{"action": "block", "reason": "..."}  // 阻止
{"action": "modify", "args": "..."}   // 修改参数（仅 pre_tool_use）
```

## 实现

- `internal/hook/hook.go` — Executor、Load、RunPre、RunPost、RunUserPrompt
- `internal/tool/runtime.go` — 在 `Execute()` 中调用 hooks
- `internal/agent/agent.go` — 在 `ChatWithStatus()` 中调用 UserPromptSubmit hook
