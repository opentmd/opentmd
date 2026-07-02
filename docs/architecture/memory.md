# 持久记忆系统 (Memory)

支持跨会话持久化用户偏好和项目事实。

## 存储位置

| 作用域 | 路径 | 说明 |
|--------|------|------|
| **全局** | `~/.opentmd/memory.md` | 用户偏好，跨项目共享 |
| **项目** | `<root>/.opentmd/memory.md` | 项目相关约定，跟随项目 |

## 文件格式

纯 `- ` 子弹列表格式，人工可编辑、git-diffable：

```markdown
- prefers tabs over spaces
- use pnpm for package management
- test before pushing
```

## 自动注入

每轮对话开始时，合并的 memory 内容会注入到 system prompt：

```
=== MEMORY ===
The user has asked you to remember these facts and preferences:

[Global]
- prefers tabs over spaces

[Project: myproj]
- use pnpm only
```

超过 4000 字符会自动截断，并提示 `[...truncated, run /memory to review]`。

## TUI 命令

| 命令 | 说明 |
|------|------|
| `/remember <text>` | 记忆一条全局事实 |
| `/remember project <text>` | 记忆一条项目事实 |
| `/forget <keyword>` | 删除包含 keyword 的全局记忆 |
| `/forget project <keyword>` | 删除包含 keyword 的项目记忆 |
| `/memory` | 查看所有记忆（全局 + 项目） |

## 实现

当前实现位于 `internal/memory/`：

| 文件 | 说明 |
|------|------|
| `store.go` | Store — Load / Append / RemoveMatching / FindMatching / MergedForPrompt |
| `store_test.go` | 单元测试 |

## 相关

- TUI 命令详见 [使用指南](../user-guide/usage.md#持久记忆)
- 对话压缩见 [Compaction](compaction.md)（独立模块，不修改 memory.md）
