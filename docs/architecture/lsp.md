# LSP 诊断系统

集成 Language Server Protocol（LSP）实现实时诊断。

## 支持的语言服务器

| 语言 | LSP 服务器 | 检测文件 |
|------|-----------|---------|
| Go | `gopls` | go.mod |
| Rust | `rust-analyzer` | Cargo.toml |
| TypeScript/JavaScript | `typescript-language-server` | tsconfig.json, package.json |
| Python | `pylsp` | pyproject.toml, requirements.txt |
| Java | `jdtls` | pom.xml |
| C/C++ | `clangd` | .clangd, compile_commands.json |

## 工具

| 工具 | 说明 |
|------|------|
| `diagnostics` | 获取当前文件/全部文件的 LSP 诊断结果 |
| 可选参数 `path` 指定文件，`severity` 过滤（error/warning/all） |

## TUI 命令

```bash
/lsp              # 显示 LSP 状态
/lsp reload       # 重启 LSP 客户端
```

## 实现

- `internal/lsp/client.go` — LSP JSON-RPC 客户端
- `internal/lsp/manager.go` — LSP 生命周期管理
- `internal/lsp/registry.go` — LSP 服务器注册与自动检测
- `internal/lsp/status.go` — 状态报告
- `internal/diagnostics/diagnostics.go` — 诊断收集
