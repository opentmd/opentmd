# 安装指南

## 系统要求

| 项目 | 要求 |
|------|------|
| 操作系统 | Linux / macOS / Windows |
| 架构 | amd64 / arm64（Windows 仅 x64） |
| Node.js | ≥ 18（仅 npm 安装时需要） |
| Go | ≥ 1.22（仅源码编译时需要） |

## 安装方式

### 方式一：curl 一键安装（推荐，Linux / macOS）

无需预先安装 Go 或 Node.js，自动检测平台并下载 GitHub Release 预编译二进制：

> Windows 用户请使用下方的 **npm 全局安装** 或 **Go 直接编译**。

```bash
curl -fsSL https://raw.githubusercontent.com/opentmd/opentmd-cli/master/scripts/install.sh | bash
```

指定版本或安装目录：

```bash
curl -fsSL .../install.sh | bash -s -- --version v0.1.0
OPENTMD_INSTALL_DIR=~/.local/bin ./scripts/install.sh
```

安装脚本会：

1. 从 GitHub API 自动检测 latest release（可用 `OPENTMD_VERSION` 覆盖）
2. 使用 `curl` 或 `wget` 下载对应平台二进制
3. 安装到 `/usr/local/bin` 或 `~/.local/bin`
4. 自动将安装目录追加到 `~/.bashrc` / `~/.zshrc`（可用 `OPENTMD_NO_PATH_EDIT=1` 跳过）
5. 执行 `opentmd --version` 验证安装

下载失败时，若在源码目录内运行，会自动回退为 `go build` 编译安装。

### 方式二：npm 全局安装

```bash
npm install -g @opentmd/cli
```

npm 会根据 `process.platform` + `process.arch` 自动下载匹配平台的预编译二进制（通过 optionalDependencies 平台包）。支持 **Windows x64**、macOS（arm64/x64）、Linux（arm64/x64）。

要求 Node.js ≥ 18。在 Git Bash、PowerShell 或 CMD 中均可使用：

```powershell
npm install -g @opentmd/cli
```

安装后直接使用：

```bash
opentmd
opentmd --version
```

### 方式三：本地源码安装

```bash
git clone https://github.com/opentmd/opentmd-cli.git
cd opentmd-cli
./scripts/install.sh
```

### 方式四：手动编译

```bash
./scripts/build.sh
# 输出: bin/opentmd

# 或仓库内快捷启动（缺失时自动 build）
./opentmd --help
```

### 方式五：Go 直接编译

```bash
CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.Version=dev" -o bin/opentmd ./cmd/opentmd/
```

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `OPENTMD_VERSION` | Release tag | 自动检测 latest |
| `OPENTMD_REPO` | GitHub 仓库 | `opentmd/opentmd-cli` |
| `OPENTMD_INSTALL_DIR` | 安装目录 | 自动检测 |
| `OPENTMD_NO_PATH_EDIT` | 设为 `1` 跳过 shell rc PATH 写入 | — |
| `USE_CN_MIRROR` | 源码编译时启用国内 Go 代理 | `1` |

## 验证安装

```bash
opentmd --version
opentmd --help
opentmd config
```

## 卸载

详见 [卸载指南](uninstall.md)。

```bash
# 交互式卸载
curl -fsSL https://raw.githubusercontent.com/opentmd/opentmd-cli/master/scripts/uninstall.sh | bash

# npm 安装的用户
npm uninstall -g @opentmd/cli
```

## 下一步

```bash
opentmd login              # 配置 Provider 与 API Key
cd your-project && opentmd  # 进入 TUI
```
