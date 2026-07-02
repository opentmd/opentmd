# 对话压缩 (Compaction)

参考 AtomCode 的 `/compact` 命令，将会话中较早的消息合并为 LLM 摘要，减少上下文占用。

## 触发方式

TUI 斜杠命令：

```bash
/compact              # 自动摘要旧消息
/compact LSP 集成     # 聚焦某主题的摘要
```

## 行为

1. 会话至少 **8 条消息**才触发压缩
2. 保留最近 **6 条**消息完整内容
3. 对更早的消息生成 LLM 摘要
4. 合并为一条 `[Previous conversation summary]` 用户消息写入 session
5. 若摘要未减小上下文体积，自动回滚

LLM 摘要失败时回退为机械拼接摘要。

## 实现

| 文件 | 说明 |
|------|------|
| `internal/compaction/compaction.go` | PlanSession / Summarize / Apply |
| `internal/agent/compact.go` | `Agent.Compact()` |
| `internal/session/session.go` | `ReplaceMessages()` 持久化 |

## 与 AtomCode 的差异

AtomCode 在完整对话（含 tool result）上做 cache-friendly stub + drain+summarize 多层策略。OpenTMD 当前 session 仅持久化 user/assistant 消息，压缩针对 session 级别的 user/assistant 轮次，实现更轻量。
