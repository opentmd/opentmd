#!/usr/bin/env bash
# opentmd-cli 卸载脚本
#
# 用法:
#   curl -fsSL https://raw.githubusercontent.com/opentmd/opentmd/master/scripts/uninstall.sh | bash
#   或本地: ./scripts/uninstall.sh
#
# 环境变量:
#   OPENTMD_REMOVE_CONFIG  设为 1 删除 ~/.opentmd 配置目录

set -euo pipefail

BINARY_NAME="opentmd"

log() { printf '==> %s\n' "$*"; }
warn() { printf '!! %s\n' "$*" >&2; }

find_binary() {
  if command -v "$BINARY_NAME" &>/dev/null; then
    command -v "$BINARY_NAME"
    return 0
  fi
  for dir in /usr/local/bin ~/.local/bin /usr/bin /opt/homebrew/bin; do
    if [ -x "${dir}/${BINARY_NAME}" ]; then
      echo "${dir}/${BINARY_NAME}"
      return 0
    fi
  done
  return 1
}

main() {
  log "opentmd-cli uninstall"

  BIN_PATH="$(find_binary || true)"
  if [ -z "$BIN_PATH" ]; then
    warn "未找到 $BINARY_NAME，可能未安装"
  else
    BIN_DIR="$(dirname "$BIN_PATH")"
    rm -f "$BIN_PATH"
    log "removed: $BIN_PATH"
  fi

  if [ "${OPENTMD_REMOVE_CONFIG:-}" = "1" ]; then
    CONFIG_DIR="${HOME}/.opentmd"
    if [ -d "$CONFIG_DIR" ]; then
      rm -rf "$CONFIG_DIR"
      log "removed config dir: $CONFIG_DIR"
    fi
  else
    CONFIG_DIR="${HOME}/.opentmd"
    if [ -d "$CONFIG_DIR" ]; then
      warn "保留配置目录: $CONFIG_DIR"
      warn "如需删除，请手动运行: rm -rf $CONFIG_DIR"
      warn "或设置 OPENTMD_REMOVE_CONFIG=1 重新运行"
    fi
  fi

  log "卸载完成"
}

main "$@"
