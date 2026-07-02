# 发布指南

## GitHub Release

### 1. 交叉编译

```bash
VERSION=0.1.0 ./scripts/build.sh --release
ls -lh dist/
```

产物命名：`opentmd_linux_amd64`、`opentmd_linux_arm64`、`opentmd_darwin_amd64`、`opentmd_darwin_arm64`

### 2. 打包（可选）

安装脚本支持 `.tar.gz` 和裸二进制两种格式：

```bash
cd dist
for f in opentmd_*; do
  tar czf "${f}.tar.gz" "$f"
done
```

### 3. 创建 GitHub Release

上传以下资产（tag 如 `v0.1.0`）：

| 资产 | 用途 |
|------|------|
| `opentmd_linux_amd64.tar.gz` | curl 安装 / npm |
| `opentmd_linux_arm64.tar.gz` | curl 安装 / npm |
| `opentmd_darwin_amd64.tar.gz` | curl 安装 / npm |
| `opentmd_darwin_arm64.tar.gz` | curl 安装 / npm |

安装脚本会依次尝试多种 URL 模式，兼容裸二进制命名。

### 4. 验证 curl 安装

```bash
curl -fsSL .../install.sh | bash -s -- --version v0.1.0
opentmd --version
```

## npm 发布

### 包结构

| 包名 | 说明 |
|------|------|
| `@opentmd/cli` | 核心包（Node.js shim + optionalDependencies） |
| `@opentmd/cli-darwin-arm64` | macOS Apple Silicon 二进制 |
| `@opentmd/cli-darwin-x64` | macOS Intel 二进制 |
| `@opentmd/cli-linux-arm64` | Linux ARM64 二进制 |
| `@opentmd/cli-linux-x64` | Linux AMD64 二进制 |

源码位于 `packages/npm/`。

### 发布命令

需先完成 GitHub Release 上传（npm 脚本从 Release 下载二进制）：

```bash
./packages/npm/scripts/build.sh 0.1.0

# 预演
./packages/npm/scripts/build.sh 0.1.0 --dry-run
```

### 用户安装

```bash
npm install -g @opentmd/cli
opentmd --version
```

## 安装脚本分发

用户可通过以下方式安装：

```bash
# curl（无需 Node.js）
curl -fsSL https://raw.githubusercontent.com/opentmd/opentmd-cli/master/scripts/install.sh | bash

# npm（需 Node.js ≥ 18）
npm install -g @opentmd/cli

# 源码
git clone ... && ./scripts/install.sh
```

## 版本号约定

- CLI `--version` 由构建时 `-ldflags "-X main.Version=..."` 注入
- npm 包版本与 GitHub Release tag 对齐（不含 `v` 前缀或一致即可，发布脚本自动处理）
- 开发构建默认为 `dev`
