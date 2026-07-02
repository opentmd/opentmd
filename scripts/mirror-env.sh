#!/usr/bin/env bash
# 国内镜像环境变量（被 scripts/build.sh / install.sh source）
#
# 默认启用国内源。关闭: USE_CN_MIRROR=0 ./scripts/build.sh
# 自定义 Go 代理:  GOPROXY=https://your-proxy,direct ./scripts/build.sh

: "${USE_CN_MIRROR:=1}"

if [ "$USE_CN_MIRROR" != "1" ]; then
  return 0 2>/dev/null || exit 0
fi

# Go module proxy（goproxy.cn 优先，失败时回退 goproxy.io，最后直连）
export GOPROXY="${GOPROXY:-https://goproxy.cn,https://goproxy.io,direct}"
# 校验库在国内常超时；USE_CN_MIRROR=1 时强制关闭（忽略 shell/go env 里的 GOSUMDB）
# 如需校验: USE_CN_MIRROR=0 GOSUMDB=sum.golang.org ./scripts/build.sh
export GOSUMDB=off

# npm（VS Code 扩展构建）
export NPM_CONFIG_REGISTRY="${NPM_CONFIG_REGISTRY:-https://registry.npmmirror.com}"
