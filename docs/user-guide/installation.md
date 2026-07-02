# 安装指南

## 系统要求

- **操作系统**: Linux / macOS
- **架构**: amd64 / arm64
- **依赖**: Go 1.22+ (源码编译)

## 安装方式

### 方式一：远程安装（推荐）

```bash
curl -fsSL https://raw.githubusercontent.com/opentmd/opentmd/master/scripts/install.sh | bash
```

安装到 `/usr/local/bin` 或 `~/.local/bin`。

### 方式二：本地源码安装

```bash
git clone <repo-url>
cd opentmd-cli
./scripts/install.sh
```

### 方式三：手动编译

```bash
git clone <repo-url>
cd opentmd-cli
./scripts/build.sh
# 输出到 bin/opentmd
```

### 方式四：Go 直接编译

```bash
go build -o ~/.local/bin/opentmd ./cmd/opentmd/
```

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `OPENTMD_VERSION` | 安装版本 | latest |
| `OPENTMD_REPO` | GitHub 仓库 | opentmd/opentmd |
| `OPENTMD_INSTALL_DIR` | 安装目录 | 自动检测 |

## 下一步

安装完成后：

```bash
# 1. 配置 Provider
opentmd login

# 2. 进入项目目录运行
cd your-project && opentmd
```
