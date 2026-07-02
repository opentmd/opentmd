#!/usr/bin/env bash
# opentmd-cli 本地部署：构建并替换本机 opentmd 二进制，便于快速验证
#
# 用法:
#   ./scripts/deploy-local.sh              # build → 替换 PATH 中的 opentmd
#   ./scripts/deploy-local.sh --test       # 先跑测试再 build + 部署
#   ./scripts/deploy-local.sh --vscode     # 同时编译 VS Code 扩展
#
# 环境变量:
#   OPENTMD_INSTALL_DIR  目标目录 (default: 自动检测，见下方)
#   BUILD_OUTPUT         构建产物路径 (default: <repo>/bin/opentmd)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
# shellcheck source=scripts/mirror-env.sh
source "${SCRIPT_DIR}/mirror-env.sh"

BINARY_NAME="opentmd"
BUILD_OUTPUT="${BUILD_OUTPUT:-${ROOT}/bin/opentmd}"

BUILD_ARGS=()

log() { printf '==> %s\n' "$*"; }
warn() { printf '!! %s\n' "$*" >&2; }
die() { printf '!! %s\n' "$*" >&2; exit 1; }

usage() {
  cat <<'EOF'
opentmd-cli local deploy

Usage: ./scripts/deploy-local.sh [build options]

Build options (passed to ./scripts/build.sh):
  -t, --test       构建前运行 go test ./...
  -e, --vscode     编译 extensions/vscode-opentmd
  -a, --all        --test + CLI + --vscode
  -h, --help       显示帮助

Environment:
  OPENTMD_INSTALL_DIR   部署目标目录 (default: PATH 中已有 opentmd，或 ~/.local/bin)
  BUILD_OUTPUT          构建产物路径 (default: bin/opentmd)

Examples:
  ./scripts/deploy-local.sh
  ./scripts/deploy-local.sh --test
  OPENTMD_INSTALL_DIR=/usr/local/bin ./scripts/deploy-local.sh
EOF
}

while [ $# -gt 0 ]; do
  case "$1" in
    -h|--help)
      usage
      exit 0
      ;;
    -t|--test|-e|--vscode|-a|--all)
      BUILD_ARGS+=("$1")
      ;;
    *)
      die "unknown option: $1 (try --help)"
      ;;
  esac
  shift
done

choose_install_dir() {
  if [ -n "${OPENTMD_INSTALL_DIR:-}" ]; then
    echo "${OPENTMD_INSTALL_DIR}"
    return
  fi
  if [ -w /usr/local/bin ]; then
    echo "/usr/local/bin"
    return
  fi
  mkdir -p "${HOME}/.local/bin"
  echo "${HOME}/.local/bin"
}

resolve_install_target() {
  if [ -n "${OPENTMD_INSTALL_DIR:-}" ]; then
    echo "${OPENTMD_INSTALL_DIR}/${BINARY_NAME}"
    return
  fi

  local existing=""
  if command -v "${BINARY_NAME}" >/dev/null 2>&1; then
    existing="$(command -v "${BINARY_NAME}")"
    if [ -f "$existing" ] || [ -L "$existing" ]; then
      echo "$existing"
      return
    fi
  fi

  echo "$(choose_install_dir)/${BINARY_NAME}"
}

deploy_binary() {
  local target="$1"
  local dir

  [ -f "${BUILD_OUTPUT}" ] || die "build output not found: ${BUILD_OUTPUT}"
  dir="$(dirname "$target")"
  mkdir -p "$dir"
  install -m 755 "${BUILD_OUTPUT}" "$target"
}

main() {
  log "opentmd-cli local deploy"
  log "build: ${SCRIPT_DIR}/build.sh ${BUILD_ARGS[*]:-}"
  "${SCRIPT_DIR}/build.sh" "${BUILD_ARGS[@]}"

  local target
  target="$(resolve_install_target)"
  log "deploy: ${BUILD_OUTPUT} → ${target}"
  deploy_binary "$target"

  log "done: ${target}"
  ls -lh "$target"

  case ":${PATH}:" in
    *":$(dirname "$target"):"*) ;;
    *)
      warn "目标目录不在 PATH 中，请添加:"
      warn "  export PATH=\"$(dirname "$target"):\$PATH\""
      ;;
  esac

  echo ""
  echo "验证: opentmd --help"
  echo ""
}

main "$@"
