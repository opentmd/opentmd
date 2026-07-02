# 会话管理系统

## 概述

会话系统位于 `internal/session/`，提供多轮对话的记忆持久化和恢复功能。

## 核心类型

### Message

```go
type Message struct {
    Role      llm.Role
    Content   string
    Timestamp time.Time
}
```

### Session

```go
type Session struct {
    ID        string
    Title     string
    CreatedAt time.Time
    UpdatedAt time.Time
    Messages  []Message
}
```

### Store

```go
type Store struct {
    dir     string   // 存储目录 (~/.opentmd/sessions/)
    persist bool     // 是否启用持久化
    current *Session // 当前活跃会话
}
```

## 存储格式

每个 Session 存储为独立的 JSON 文件：

```bash
~/.opentmd/sessions/
├── <uuid>.json
└── <uuid>.json
```

文件内容示例：

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "title": "New Session",
  "created_at": "2025-01-01T00:00:00Z",
  "updated_at": "2025-01-01T00:10:00Z",
  "messages": [
    {
      "role": "user",
      "content": "你好",
      "timestamp": "2025-01-01T00:00:00Z"
    },
    {
      "role": "assistant",
      "content": "你好！",
      "timestamp": "2025-01-01T00:00:10Z"
    }
  ]
}
```

## 关键功能

### 启动恢复

`NewStore()` 启动时自动恢复最新 Session：
1. 列出所有会话（按更新时间降序）
2. 有历史 → 加载最新会话
3. 无历史 → 新建会话

### 消息管理

| 方法 | 功能 |
|------|------|
| `AddMessage()` | 添加消息（自动保存） |
| `ProviderMessages()` | 获取 Provider 格式的消息列表（限最近 50 条，防止上下文超长） |
| `Clear()` | 清空当前会话 |
| `New()` | 新建会话 |
| `Load(id)` | 加载指定会话 |
| `List()` | 列出所有会话 |
| `saveCurrent()` | 持久化当前会话（仅 `persist=true` 时） |

### 上下文窗口限制

`ProviderMessages()` 截取最近 50 条消息，防止上下文过长。

## 配置关联

在 `config.toml` 中控制：

```toml
[session]
persist = true   # 是否持久化会话到磁盘
```

- `persist: true` → 每次 `AddMessage()` 自动写入磁盘
- `persist: false` → 仅内存中保存，重启丢失
