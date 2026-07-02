#!/usr/bin/env bash
# 一键配置 Go / npm 国内镜像（写入用户级配置，持久生效）
#
# 用法: ./scripts/setup-mirror.sh

set -euo pipefail

log() { printf '==> %s\n' "$*"; }

GOPROXY_URL="https://goproxy.cn,https://goproxy.io,direct"
GOSUMDB_URL="${GOSUMDB_URL:-off}"
NPM_REGISTRY="https://registry.npmmirror.com"

if command -v go >/dev/null 2>&1; then
  log "Go GOPROXY → ${GOPROXY_URL}"
  go env -w "GOPROXY=${GOPROXY_URL}"
  log "Go GOSUMDB → ${GOSUMDB_URL}"
  go env -w "GOSUMDB=${GOSUMDB_URL}"
  go env GOPROXY GOSUMDB
else
  echo "!! go not found, skip Go mirror setup" >&2
fi

if command -v npm >/dev/null 2>&1; then
  log "npm registry → ${NPM_REGISTRY}"
  npm config set registry "$NPM_REGISTRY"
  npm config get registry
else
  echo "!! npm not found, skip npm mirror setup" >&2
fi

log "done"
echo ""
echo "若 sumdb 可访问，可额外执行:"
echo "  go env -w GOSUMDB=sum.golang.google.cn"
echo ""
echo "单次构建（不写入全局）:"
echo "  ./scripts/build.sh --all"
