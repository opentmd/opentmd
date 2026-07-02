# 权限模型 (Permission)

参考 AtomCode 的 `turn/permission.rs` 和 `tool/mod.rs` 权限系统实现。

## 权限级别

| 级别 | 常量 | 说明 |
|------|------|------|
| 自动允许 | `AutoApprove` | 读操作自动执行（read_file, grep, glob, web_search 等） |
| 需审批 | `RequireApproval` | 写操作需用户确认（write_file, edit_file, bash, MCP 工具） |
| 总需审批 | `RequireApprovalAlways` | 高危操作每次均需确认（写文件到工作区外、危险 shell 命令） |

## 权限分类规则

```
read_file, edit_file, grep, glob, list_directory  → AutoApprove
write_file                                          → RequireApproval
edit_file, search_replace                           → RequireApproval
bash, run_shell                                     → RequireApproval
mcp__* (外部 MCP 工具)                               → RequireApproval
```

## 危险 shell 检测

以下模式会被识别为 `RequireApprovalAlways`：

`rm -rf`, `rm -fr`, `sudo `, `mkfs`, `dd if=`, `:(){`, `chmod -r`,
`> /dev/sd`, `shutdown`, `reboot`, `curl | sh`, `wget | sh`

## 审批会话流

```
工具执行请求
    ↓
权限检查 (Check)
    ├─ AutoApprove → 直接执行
    ├─ RequireApproval
    │     ├─ 已 session grant → 直接执行
    │     └─ 未 grant → 请求用户审批
    │           ├─ Y/y → AllowOnce（仅本次）
    │           ├─ A/a → AllowSession（本会话内）
    │           └─ N/n → Deny
    └─ RequireApprovalAlways → 请求用户审批（不受 session grant 绕过）
```

## 实现

- `internal/permission/permission.go` — Store、Classify、Check、决策枚举
- `internal/tool/runtime.go:Execute()` — 权限钩子集成

参考 AtomCode `docs/atomcode/crates/atomcode-core/src/turn/permission.rs`。
