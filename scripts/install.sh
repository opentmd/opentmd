#!/usr/bin/env bash
# opentmd-cli 一键安装脚本
#
# 用法:
#   curl -fsSL https://raw.githubusercontent.com/opentmd/opentmd-cli/master/scripts/install.sh | bash
#   或本地: ./scripts/install.sh
#
# 环境变量:
#   OPENTMD_VERSION   版本号 (default: latest)
#   OPENTMD_REPO      GitHub 仓库 (default: opentmd/opentmd-cli)
#   OPENTMD_INSTALL_DIR  安装目录 (auto: /usr/local/bin 或 ~/.local/bin)

set -eu
if command -v bash >/dev/null 2>&1; then
  set -o pipefail 2>/dev/null || true
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=scripts/mirror-env.sh
source "${SCRIPT_DIR}/mirror-env.sh"

OPENTMD_VERSION="${OPENTMD_VERSION:-latest}"
OPENTMD_REPO="${OPENTMD_REPO:-opentmd/opentmd-cli}"
BINARY_NAME="opentmd"

log() { printf '==> %s\n' "$*"; }
warn() { printf '!! %s\n' "$*" >&2; }

detect_os() {
  local os
  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  case "$os" in
    linux) echo "linux" ;;
    darwin) echo "darwin" ;;
    *) warn "unsupported OS: $os"; return 1 ;;
  esac
}

detect_arch() {
  local arch
  arch="$(uname -m)"
  case "$arch" in
    x86_64|amd64) echo "amd64" ;;
    aarch64|arm64) echo "arm64" ;;
    *) warn "unsupported arch: $arch"; return 1 ;;
  esac
}

choose_install_dir() {
  if [ -n "${OPENTMD_INSTALL_DIR:-}" ]; then
    echo "$OPENTMD_INSTALL_DIR"
    return
  fi
  if [ -w /usr/local/bin ]; then
    echo "/usr/local/bin"
    return
  fi
  mkdir -p "${HOME}/.local/bin"
  echo "${HOME}/.local/bin"
}

install_from_release() {
  local os arch url tmpdir
  os="$(detect_os)"
  arch="$(detect_arch)"
  url="https://github.com/${OPENTMD_REPO}/releases/download/${OPENTMD_VERSION}/${BINARY_NAME}_${os}_${arch}.tar.gz"

  log "download ${url}"
  tmpdir="$(mktemp -d)"

  if ! curl -fsSL --connect-timeout 5 --max-time 30 "$url" -o "${tmpdir}/release.tar.gz" 2>/dev/null; then
    rm -rf "$tmpdir"
    return 1
  fi
  tar -xzf "${tmpdir}/release.tar.gz" -C "$tmpdir"
  install -m 755 "${tmpdir}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
  rm -rf "$tmpdir"
  return 0
}

install_from_source() {
  local root
  root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
  if [ ! -f "${root}/go.mod" ]; then
    return 1
  fi
  log "build from source (${root})"
  log "GOPROXY=${GOPROXY:-default}"
  (
    cd "$root"
    go mod download
    CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o "${INSTALL_DIR}/${BINARY_NAME}" ./cmd/opentmd/
  )
  return 0
}

main() {
  INSTALL_DIR="$(choose_install_dir)"
  mkdir -p "$INSTALL_DIR"

  log "opentmd-cli install"
  log "install dir: ${INSTALL_DIR}"

  if install_from_release; then
    log "installed from release"
  elif install_from_source; then
    log "installed from source"
  else
    warn "无法下载 release，且未找到源码目录"
    warn "请手动运行: go build -o ~/.local/bin/opentmd ./cmd/opentmd/"
    exit 1
  fi

  log "done: ${INSTALL_DIR}/${BINARY_NAME}"

  case ":${PATH}:" in
    *":${INSTALL_DIR}:"*) ;;
    *)
      warn "请将 ${INSTALL_DIR} 加入 PATH，例如在 ~/.bashrc 添加:"
      warn "  export PATH=\"${INSTALL_DIR}:\$PATH\""
      ;;
  esac

  echo ""
  echo "下一步:"
  echo "  1. opentmd login              # 配置 Provider"
  echo "  2. cd your-project && opentmd"
  echo ""
}

main "$@"
