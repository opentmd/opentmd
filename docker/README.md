# OpenTMD Docker 镜像

本目录包含 OpenTMD Daemon 的 Docker 镜像。

## 构建

```bash
# 先构建 Go 二进制
go build -o docker/opentmd ./cmd/opentmd/

# 构建 Docker 镜像
docker build -t opentmd-daemon -f docker/Dockerfile-daemon docker/
```

## 运行

```bash
# 基本运行
docker run -d --name opentmd-daemon \
  -p 13456:13456 \
  opentmd-daemon

# 挂载配置和项目
docker run -d --name opentmd-daemon \
  -p 13456:13456 \
  -v ~/.opentmd:/root/.opentmd \
  -v $(pwd):/workspace \
  opentmd-daemon

# 传递 API Key 环境变量
docker run -d --name opentmd-daemon \
  -p 13456:13456 \
  -e DEEPSEEK_API_KEY=your-key \
  -v ~/.opentmd:/root/.opentmd \
  opentmd-daemon
```

## 验证

```bash
curl http://localhost:13456/health
```
