# 构建指南

## 系统要求

- Go 1.22+
- Linux 或 macOS（交叉编译支持 darwin）

## 快速构建

```bash
./scripts/build.sh
# 输出: bin/opentmd

# 仓库内快捷启动
./opentmd --help
```

## 构建选项

```bash
./scripts/build.sh --test       # 构建前跑 go test ./...
./scripts/build.sh --vscode     # 同时编译 VS Code 扩展
./scripts/build.sh --release    # 交叉编译 → dist/
./scripts/build.sh --all        # test + cli + vscode
```

## 自定义构建

```bash
# 指定输出路径与版本
OUTPUT=./my-opentmd VERSION=0.1.0 ./scripts/build.sh

# 直接使用 go build
CGO_ENABLED=0 go build \
  -trimpath \
  -ldflags "-s -w -X main.Version=0.1.0" \
  -o bin/opentmd ./cmd/opentmd/
```

构建参数说明：

- `CGO_ENABLED=0` — 静态编译，无 C 依赖
- `-trimpath` — 去除源码路径
- `-ldflags "-s -w -X main.Version=..."` — 精简体积并注入版本号

## 本地部署

将构建产物替换到 PATH 中的 `opentmd`：

```bash
./scripts/deploy-local.sh
./scripts/deploy-local.sh --test
```

## 国内镜像

默认启用国内 Go/npm 镜像（`USE_CN_MIRROR=1`）。关闭：

```bash
USE_CN_MIRROR=0 ./scripts/build.sh
```

详见 `scripts/mirror-env.sh`。

## 发布流程

完整发布步骤见 [发布指南](release.md)。

简要流程：

1. `VERSION=0.1.0 ./scripts/build.sh --release` → `dist/opentmd_*`
2. 打包上传到 GitHub Release
3. `curl | bash` 安装脚本自动拉取
4. `./packages/npm/scripts/build.sh 0.1.0` 发布 npm 包

## 文档站点

```bash
cd docs/site && npm install && npm run build && npm run serve
```

线上地址：[www.opentmd.com](https://www.opentmd.com) · 配置见 [docs/site/DNS.md](../site/DNS.md)

## 测试

```bash
go test ./...
./scripts/build.sh --test
```

详见 [测试指南](testing.md)。
