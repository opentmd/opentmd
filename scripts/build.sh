#!/usr/bin/env bash
# opentmd-cli 构建脚本
#
# 用法:
#   ./scripts/build.sh              # 构建 CLI → bin/opentmd
#   ./scripts/build.sh --test       # 构建前先跑测试
#   ./scripts/build.sh --vscode     # 同时编译 VS Code 扩展
#   ./scripts/build.sh --release    # 交叉编译发布包 → dist/
#   ./scripts/build.sh --all        # test + cli + vscode
#   ./scripts/deploy-local.sh       # 构建并替换本机 PATH 中的 opentmd
#   ./opentmd                       # 仓库内快捷启动（缺失时自动 build）
#
# 环境变量:
#   OUTPUT   CLI 输出路径 (default: bin/opentmd)
#   DIST     发布目录 (default: dist)
#   VERSION  发布版本标签，用于 dist 文件名 (default: dev)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "$ROOT"

# 国内 Go/npm 镜像（USE_CN_MIRROR=0 可关闭）
# shellcheck source=scripts/mirror-env.sh
source "${SCRIPT_DIR}/mirror-env.sh"

OUTPUT="${OUTPUT:-bin/opentmd}"
DIST="${DIST:-dist}"
VERSION="${VERSION:-dev}"

DO_TEST=0
DO_VSCODE=0
DO_RELEASE=0
DO_CLI=1

log() { printf '==> %s\n' "$*"; }
die() { printf '!! %s\n' "$*" >&2; exit 1; }

usage() {
  cat <<'EOF'
opentmd-cli build

Usage: ./scripts/build.sh [options]

Options:
  -t, --test       构建前运行 go test ./...
  -e, --vscode     编译 extensions/vscode-opentmd
  -r, --release    交叉编译 linux/darwin × amd64/arm64 → dist/
  -a, --all        --test + CLI + --vscode
  -h, --help       显示帮助

Environment:
  OUTPUT=bin/opentmd   CLI 输出路径
  DIST=dist            发布目录
  VERSION=dev          发布包版本号
  USE_CN_MIRROR=1    使用国内 Go/npm 镜像 (default: 1)
  GOPROXY=...          自定义 Go 代理

Examples:
  ./scripts/build.sh
  OUTPUT=/usr/local/bin/opentmd ./scripts/build.sh
  ./scripts/build.sh --test --vscode
  VERSION=0.2.0 ./scripts/build.sh --release
EOF
}

while [ $# -gt 0 ]; do
  case "$1" in
    -t|--test)    DO_TEST=1 ;;
    -e|--vscode)  DO_VSCODE=1 ;;
    -r|--release) DO_RELEASE=1; DO_CLI=0 ;;
    -a|--all)     DO_TEST=1; DO_VSCODE=1 ;;
    -h|--help)    usage; exit 0 ;;
    *) die "unknown option: $1 (try --help)" ;;
  esac
  shift
done

require_go() {
  command -v go >/dev/null 2>&1 || die "go not found; install Go 1.22+"
  if [ "${USE_CN_MIRROR:-1}" = "1" ]; then
    log "GOPROXY=${GOPROXY}"
    log "GOSUMDB=${GOSUMDB}"
  fi
}

run_tests() {
  log "go test ./..."
  go test ./...
}

build_cli() {
  log "opentmd-cli build"
  log "output: ${OUTPUT}"
  mkdir -p "$(dirname "$OUTPUT")"
  log "go mod tidy"
  go mod tidy
  log "go build"
  CGO_ENABLED=0 go build \
    -trimpath \
    -ldflags "-s -w" \
    -o "$OUTPUT" \
    ./cmd/opentmd/
  log "done: ${OUTPUT}"
}

build_vscode() {
  local ext="${ROOT}/extensions/vscode-opentmd"
  [ -f "${ext}/package.json" ] || die "VS Code extension not found: ${ext}"
  command -v npm >/dev/null 2>&1 || die "npm not found; required for --vscode"

  log "VS Code extension build (${ext})"
  (
    cd "$ext"
    if [ -f package-lock.json ]; then
      npm ci --silent
    else
      npm install --silent
    fi
    npm run compile --silent
  )
  log "done: ${ext}/out/extension.js"
}

build_release() {
  require_go
  mkdir -p "$DIST"
  log "release build → ${DIST}/ (version=${VERSION})"
  log "go mod tidy"
  go mod tidy

  local targets=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
  )

  for target in "${targets[@]}"; do
    local goos="${target%%/*}"
    local goarch="${target##*/}"
    local name="opentmd_${goos}_${goarch}"
    [ "$VERSION" != "dev" ] && name="opentmd_${VERSION}_${goos}_${goarch}"
    local out="${DIST}/${name}"
    log "  ${goos}/${goarch} → ${out}"
    CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" go build \
      -trimpath \
      -ldflags "-s -w" \
      -o "$out" \
      ./cmd/opentmd/
    chmod +x "$out"
  done

  log "done: ${DIST}/"
  ls -lh "$DIST"/opentmd_* 2>/dev/null || true
}

main() {
  require_go

  if [ "$DO_TEST" -eq 1 ]; then
    run_tests
  fi

  if [ "$DO_RELEASE" -eq 1 ]; then
    build_release
  fi

  if [ "$DO_CLI" -eq 1 ]; then
    build_cli
  fi

  if [ "$DO_VSCODE" -eq 1 ]; then
    build_vscode
  fi
}

main
