# 发布指南

## GitHub Release（自动）

推送 `v*` tag 后，`.github/workflows/release.yml` 自动执行：

1. 交叉编译 linux / darwin / windows（amd64/arm64，Windows 仅 amd64）
2. 创建 GitHub Release 并上传资产
3. 使用 Environment `Configure NPM` 中的 `NPM_TOKEN` 发布 npm 包

```bash
git tag v0.1.1
git push origin v0.1.1
```

### GitHub 环境配置

在 **Settings → Environments → Configure NPM** 中：

| 项 | 配置 |
|----|------|
| **Environment secrets** | 名称 `NPM_TOKEN`，值为 npm automation token（bypass 2FA） |
| **Deployment branches and tags** | 选 **No restriction**，或添加 tag 规则 `v*` |

> 常见失败原因：secret 已添加，但 **Deployment branches and tags** 仅允许 `main` 等分支，**tag 发布（`v0.1.4`）无法读取 environment secrets**，表现为 `NPM_TOKEN is empty`。

备选：在 **Settings → Secrets and variables → Actions → Repository secrets** 添加同名 `NPM_TOKEN`（不依赖 environment 分支规则，推荐与 environment secret 二选一或同时配置）。

> 日志里 token 显示为空是正常的（GitHub 脱敏）。

## GitHub Release（手动）

### 1. 交叉编译

```bash
VERSION=0.1.0 ./scripts/build.sh --release
ls -lh dist/
```

产物命名：`opentmd_linux_amd64`、`opentmd_linux_arm64`、`opentmd_darwin_amd64`、`opentmd_darwin_arm64`、`opentmd_windows_amd64.exe`

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
| `opentmd_windows_amd64.exe` | npm（Windows x64） |

安装脚本会依次尝试多种 URL 模式，兼容裸二进制命名。Windows 仅通过 npm 分发（curl 安装脚本不支持 Windows）。

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
| `@opentmd/cli-win32-x64` | Windows x64 二进制 |

源码位于 `packages/npm/`。

### 发布命令

推送 tag 后由 GitHub Actions 自动发布。本地手动发布（需 `NPM_TOKEN`）：

```bash
export NODE_AUTH_TOKEN="$NPM_TOKEN"
./packages/npm/scripts/build.sh 0.1.0

# 预演
./packages/npm/scripts/build.sh 0.1.0 --dry-run
```

CI 工作流在 GitHub Release 创建完成后，从 Release 资产下载二进制并发布全部 6 个 npm 包。

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
