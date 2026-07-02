# 文件 I/O 模块

## 概述

文件 I/O 模块位于 `internal/fileio/`，提供安全的文件读取能力，支持大文件截断。

## 核心函数

### `Read()`

```go
func Read(path string, maxBytes int) (string, error)
```

读取文件内容，可选大小限制：

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `path` | - | 文件路径（相对或绝对） |
| `maxBytes` | 64KB | 最大读取字节数 |

### `ValidatePath()`

```go
func ValidatePath(path string) error
```

验证文件路径合法性：
- 文件必须存在
- 不能是目录

## 安全特性

1. **路径解析** — 自动转换为绝对路径
2. **目录检查** — 拒绝读取目录
3. **大小限制** — 默认 64KB 截断
4. **空路径检查** — 返回明确错误

## 使用示例

```go
content, err := fileio.Read("./main.go", 32*1024)
if err != nil {
    // 处理错误
}
// content 包含文件内容（可能截断）
```

## 集成

由 `tool.Registry` 中的 `read_file` 工具调用。
