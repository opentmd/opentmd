#!/usr/bin/env bash
# opentmd-cli 一键安装脚本
#
#   curl -fsSL https://raw.githubusercontent.com/opentmd/opentmd-cli/master/scripts/install.sh | bash
#   curl -fsSL .../install.sh | bash -s -- --version v0.1.0
#
# 环境变量:
#   OPENTMD_VERSION      版本 tag (default: 自动检测 latest release)
#   OPENTMD_REPO         GitHub 仓库 (default: opentmd/opentmd-cli)
#   OPENTMD_INSTALL_DIR  安装目录 (default: /usr/local/bin 或 ~/.local/bin)
#   OPENTMD_NO_PATH_EDIT 设为 1 跳过 shell rc PATH 写入

set -eu
if command -v bash >/dev/null 2>&1; then
  set -o pipefail 2>/dev/null || true
fi

OPENTMD_REPO="${OPENTMD_REPO:-opentmd/opentmd-cli}"
BINARY_NAME="opentmd"
INSTALLER_MARKER="# Added by opentmd-cli installer"
REPO_API="https://api.github.com/repos/${OPENTMD_REPO}/releases/latest"
RELEASE_BASE="https://github.com/${OPENTMD_REPO}/releases/download"

log() { printf '==> %s\n' "$*"; }
warn() { printf '!! %s\n' "$*" >&2; }

maybe_source_mirror_env() {
  local script_dir
  script_dir="$(cd "$(dirname "${BASH_SOURCE[0]:-$0}")" && pwd)"
  if [ -f "${script_dir}/mirror-env.sh" ]; then
    # shellcheck source=scripts/mirror-env.sh
    source "${script_dir}/mirror-env.sh"
  elif [ "${USE_CN_MIRROR:-1}" = "1" ]; then
    export GOPROXY="${GOPROXY:-https://goproxy.cn,https://goproxy.io,direct}"
    export GOSUMDB=off
  fi
}

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
  if [ -w /usr/local/bin ] 2>/dev/null; then
    echo "/usr/local/bin"
    return
  fi
  mkdir -p "${HOME}/.local/bin"
  echo "${HOME}/.local/bin"
}

setup_download_tools() {
  if command -v curl >/dev/null 2>&1; then
    FETCH="curl -fsSL --connect-timeout 10 --max-time 30"
    DOWNLOAD="curl -fL --connect-timeout 10 --max-time 120 -o"
    return 0
  fi
  if command -v wget >/dev/null 2>&1; then
    FETCH="wget -qO- --timeout=30 --tries=2"
    DOWNLOAD="wget --show-progress -O"
    return 0
  fi
  warn "需要 curl 或 wget"
  return 1
}

resolve_version() {
  if [ -n "${OPENTMD_VERSION:-}" ]; then
    echo "$OPENTMD_VERSION"
    return 0
  fi

  log "detecting latest release"
  local tag
  tag="$($FETCH "$REPO_API" 2>/dev/null | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -n1)"
  if [ -n "$tag" ]; then
    echo "$tag"
    return 0
  fi

  warn "无法自动检测版本，使用 latest"
  echo "latest"
}

try_download_release() {
  local version="$1" os="$2" arch="$3" dest="$4"
  local urls=(
    "${RELEASE_BASE}/${version}/${BINARY_NAME}_${os}_${arch}.tar.gz"
    "${RELEASE_BASE}/${version}/${BINARY_NAME}_${version}_${os}_${arch}.tar.gz"
    "${RELEASE_BASE}/${version}/${BINARY_NAME}_${os}_${arch}"
    "${RELEASE_BASE}/${version}/${BINARY_NAME}-${version}-${os}-${arch}"
  )
  local url tmp
  for url in "${urls[@]}"; do
    log "try ${url}"
    if $DOWNLOAD "$dest" "$url" 2>/dev/null; then
      if [ -s "$dest" ] && ! head -c 4 "$dest" | grep -q '<' 2>/dev/null; then
        echo "$url"
        return 0
      fi
    fi
    rm -f "$dest"
  done
  return 1
}

install_from_release() {
  local os arch version tmpdir dest url
  os="$(detect_os)"
  arch="$(detect_arch)"
  version="$(resolve_version)"
  tmpdir="$(mktemp -d)"
  dest="${tmpdir}/artifact"

  if ! url="$(try_download_release "$version" "$os" "$arch" "$dest")"; then
    rm -rf "$tmpdir"
    return 1
  fi
  log "downloaded from ${url}"

  if file "$dest" 2>/dev/null | grep -qi 'gzip compressed'; then
    tar -xzf "$dest" -C "$tmpdir"
    rm -f "$dest"
    if [ -x "${tmpdir}/${BINARY_NAME}" ]; then
      install -m 755 "${tmpdir}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
    else
      local bin
      bin="$(find "$tmpdir" -maxdepth 2 -type f -name "$BINARY_NAME" | head -n1)"
      [ -n "$bin" ] || { rm -rf "$tmpdir"; return 1; }
      install -m 755 "$bin" "${INSTALL_DIR}/${BINARY_NAME}"
    fi
  else
    install -m 755 "$dest" "${INSTALL_DIR}/${BINARY_NAME}"
  fi

  rm -rf "$tmpdir"
  return 0
}

install_from_source() {
  local root
  root="$(cd "$(dirname "${BASH_SOURCE[0]:-$0}")/.." && pwd)"
  if [ ! -f "${root}/go.mod" ]; then
    return 1
  fi
  maybe_source_mirror_env
  log "build from source (${root})"
  log "GOPROXY=${GOPROXY:-default}"
  command -v go >/dev/null 2>&1 || { warn "go not found"; return 1; }
  (
    cd "$root"
    go mod download
    CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o "${INSTALL_DIR}/${BINARY_NAME}" ./cmd/opentmd/
  )
  return 0
}

ensure_path_in_shell_rc() {
  local rc line
  if [ "${OPENTMD_NO_PATH_EDIT:-}" = "1" ]; then
    return 0
  fi
  case ":${PATH}:" in
    *":${INSTALL_DIR}:"*) return 0 ;;
  esac

  line="export PATH=\"${INSTALL_DIR}:\$PATH\""
  if [ -n "${ZSH_VERSION:-}" ] || [ "$(basename "${SHELL:-}")" = "zsh" ]; then
    rc="${HOME}/.zshrc"
  elif [ -n "${BASH_VERSION:-}" ] || [ "$(basename "${SHELL:-}")" = "bash" ]; then
    rc="${HOME}/.bashrc"
  else
    warn "请将 ${INSTALL_DIR} 加入 PATH:"
    warn "  ${line}"
    return 0
  fi

  if [ -f "$rc" ] && grep -qF "$INSTALL_DIR" "$rc" 2>/dev/null; then
    return 0
  fi

  {
    echo ""
    echo "$INSTALLER_MARKER"
    echo "$line"
  } >>"$rc"
  log "added PATH to ${rc}"
  echo ""
  echo "立即生效请运行: source ${rc}"
}

verify_install() {
  local bin="${INSTALL_DIR}/${BINARY_NAME}"
  if [ ! -x "$bin" ]; then
    warn "binary missing: $bin"
    return 1
  fi
  if "$bin" --version >/dev/null 2>&1; then
    log "verified: $("$bin" --version)"
    return 0
  fi
  if "$bin" --help >/dev/null 2>&1; then
    log "verified: $bin --help"
    return 0
  fi
  warn "安装完成但验证失败"
  return 1
}

parse_args() {
  while [ $# -gt 0 ]; do
    case "$1" in
      --version)
        shift
        OPENTMD_VERSION="${1:-}"
        [ -n "$OPENTMD_VERSION" ] || { warn "--version requires a value"; exit 1; }
        ;;
      --version=*)
        OPENTMD_VERSION="${1#*=}"
        ;;
      --dir|--install-dir)
        shift
        OPENTMD_INSTALL_DIR="${1:-}"
        [ -n "$OPENTMD_INSTALL_DIR" ] || { warn "--dir requires a value"; exit 1; }
        ;;
      --dir=*|--install-dir=*)
        OPENTMD_INSTALL_DIR="${1#*=}"
        ;;
      -h|--help)
        cat <<EOF
opentmd-cli installer

Usage:
  curl -fsSL .../install.sh | bash
  ./scripts/install.sh [--version vX.Y.Z] [--dir ~/.local/bin]

Environment:
  OPENTMD_VERSION       Release tag (default: latest from GitHub)
  OPENTMD_REPO          GitHub repo (default: opentmd/opentmd-cli)
  OPENTMD_INSTALL_DIR   Install directory
  OPENTMD_NO_PATH_EDIT  Skip shell rc PATH edit when set to 1
EOF
        exit 0
        ;;
      *)
        warn "unknown argument: $1"
        exit 1
        ;;
    esac
    shift
  done
}

main() {
  parse_args "$@"
  setup_download_tools

  INSTALL_DIR="$(choose_install_dir)"
  mkdir -p "$INSTALL_DIR"

  log "opentmd-cli install"
  log "install dir: ${INSTALL_DIR}"

  if install_from_release; then
    log "installed from release"
  elif install_from_source; then
    log "installed from source"
  else
    warn "无法下载 release，且未找到可编译的源码目录"
    warn "请手动运行: go build -o ~/.local/bin/opentmd ./cmd/opentmd/"
    exit 1
  fi

  log "done: ${INSTALL_DIR}/${BINARY_NAME}"
  verify_install || true
  ensure_path_in_shell_rc

  echo ""
  echo "下一步:"
  echo "  1. opentmd login"
  echo "  2. cd your-project && opentmd"
  echo ""
  echo "或通过 npm 安装:"
  echo "  npm install -g @opentmd/cli"
  echo ""
}

main "$@"
