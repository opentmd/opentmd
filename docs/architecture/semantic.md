# 语义分析系统

基于 tree-sitter 的多语言代码符号分析。

## 支持的语言

| 语言 | 语法文件 |
|------|---------|
| Go | `go.scm` |
| Rust | `rust.scm` |
| Python | `python.scm` |
| JavaScript | `javascript.scm` |
| TypeScript | `typescript.scm` |
| Java | `java.scm` |
| C | `c.scm` |
| C++ | `cpp.scm` |
| C# | `csharp.scm` |
| PHP | `php.scm` |

## 工具

| 工具 | 说明 |
|------|------|
| `list_symbols` | 列出文件中所有顶级符号（函数、类、类型）及其行号 |
| `read_symbol` | 按符号名读取源码片段，支持 context_lines 参数 |

## 技术栈检测

自动检测项目技术栈（`/status` 显示）：

| 检测文件 | 标签 |
|---------|------|
| go.mod | Go |
| package.json | Node.js |
| Cargo.toml | Rust |
| pyproject.toml / requirements.txt | Python |
| pom.xml | Java |

## 实现

- `internal/semantic/language.go` — 语言注册与 grammar 加载
- `internal/semantic/queries.go` — tree-sitter 查询执行
- `internal/semantic/symbols.go` — 符号提取与格式化
- `internal/semantic/treesitter.go` — tree-sitter 封装
- `internal/semantic/queries/*.scm` — 每种语言的 tree-sitter 查询
