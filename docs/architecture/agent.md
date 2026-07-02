# Agent 编排系统

## 概述

Agent 是 OpenTMD 的核心编排器，位于 `internal/agent/`，负责接收用户输入、拼装上下文、调用模型、决策和执行工具、汇总回复的多步循环。

## 核心结构

```go
type Agent struct {
    cfg     *config.Config   // 配置引用
    prov    llm.Provider     // LLM Provider
    session *session.Store   // 会话存储
    tools   *tool.Registry   // 工具注册表
}
```

## 执行流程

```
用户输入
    ↓
加载 Session 历史消息
    ↓
拼装消息列表 [历史 + 用户输入]
    ↓
runLoop:
  ├─ provider.Chat() 发送请求
  ├─ 检测响应中的 tool_calls
  │   ├─ 有 → 逐一执行工具 → 结果加入消息 → 继续循环
  │   └─ 无 → 输出最终回复 → 退出循环
  └─ 流式输出 chunk 到 onChunk 回调
    ↓
保存用户消息 + AI 回复到 Session
    ↓
返回完整回复
```

## System Prompt

Agent 预设系统提示词，定义 AI 助手的角色和行为准则：

```text
You are OpenTMD, an AI coding assistant running in the user's terminal.
You can autonomously read files, list directories, and run shell commands
using the provided tools.
```

规则包括：
- 分析后再行动，一步只调用一个工具
- 所有路径相对于用户工作目录
- 工具失败时解释错误并尝试替代方案
- 不编造文件内容或命令输出

## 流式处理

`Chat()` 方法接受 `StreamHandler` 回调，每个流式 chunk 都会触发回调，实现实时输出。

## 非交互模式

`ChatToWriter()` 封装 `Chat()`，直接将流式输出写入 `io.Writer`（如 stdout），用于 Prompt 模式。

## Session 操作

Agent 暴露 Session 操作接口供 TUI 调用：

| 方法 | 功能 |
|------|------|
| `Session()` | 获取当前 Session |
| `ClearSession()` | 清空会话 |
| `LoadSession(id)` | 加载指定 Session |
| `ListSessions()` | 列出所有 Session |
| `NewSession()` | 新建 Session |
