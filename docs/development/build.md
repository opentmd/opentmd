# 构建指南

## 系统要求

- Go 1.22+
- Linux 或 macOS

## 快速构建

```bash
cd opentmd-cli
./scripts/build.sh
# 输出: bin/opentmd
```

## 自定义构建

```bash
# 指定输出路径
OUTPUT=./my-opentmd ./scripts/build.sh

# 或直接使用 go build
go build -o bin/opentmd -ldflags="-s -w" ./cmd/opentmd/
```

## `scripts/build.sh` 详细说明

```bash
#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT"

OUTPUT="${OUTPUT:-bin/opentmd}"

echo "==> building: ${OUTPUT}"
CGO_ENABLED=0 go build -ldflags="-s -w" -o "$OUTPUT" ./cmd/opentmd/
echo "==> done: ${OUTPUT}"
```

- `CGO_ENABLED=0` — 静态编译，无 C 依赖
- `-ldflags="-s -w"` — 去除调试信息，减小二进制体积
- 输出到 `bin/opentmd`（可配置 OUTPUT 环境变量）

## 发布流程

1. 编译各平台二进制
2. 上传到 GitHub Release
3. 用户通过安装脚本下载

安装脚本自动检测 OS/Arch 并下载对应二进制。
