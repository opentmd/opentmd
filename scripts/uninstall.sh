#!/usr/bin/env bash
# opentmd-cli 卸载脚本
#
#   curl -fsSL https://raw.githubusercontent.com/opentmd/opentmd-cli/master/scripts/uninstall.sh | bash
#
# 选项:
#   --yes         跳过交互（G1=yes, G2=no, G3=yes）
#   --purge       删除全部数据（含 credentials）
#   --keep-data   仅删除二进制与 PATH 配置
#   --dry-run     仅显示计划

set -euo pipefail

BINARY_NAME="opentmd"
LEGACY_SHORT_NAME="tmd"
INSTALLER_MARKER="# Added by opentmd-cli installer"
DATA_DIR="${HOME}/.opentmd"

GROUP2_FILES="config.toml mcp.json hooks.json"
GROUP3_FILES="memory.md OPENTMD.md"
GROUP3_DIRS="sessions plugins skills"

DO_YES=0
DO_PURGE=0
DO_KEEP=0
DO_DRY=0

log() { printf '==> %s\n' "$*"; }
warn() { printf '!! %s\n' "$*" >&2; }

parse_args() {
  while [ $# -gt 0 ]; do
    case "$1" in
      --yes) DO_YES=1 ;;
      --purge) DO_PURGE=1 ;;
      --keep-data) DO_KEEP=1 ;;
      --dry-run) DO_DRY=1 ;;
      -h|--help)
        cat <<EOF
opentmd-cli uninstall

Flags:
  --yes         non-interactive (G1=yes, G2=no, G3=yes)
  --purge       remove everything including credentials
  --keep-data   remove binary + PATH only
  --dry-run     print plan only
EOF
        exit 0
        ;;
      *)
        warn "unknown flag: $1"
        exit 2
        ;;
    esac
    shift
  done
}

find_binary() {
  if command -v "$BINARY_NAME" &>/dev/null; then
    command -v "$BINARY_NAME"
    return 0
  fi
  for dir in /usr/local/bin "${HOME}/.local/bin" /usr/bin /opt/homebrew/bin; do
    if [ -x "${dir}/${BINARY_NAME}" ]; then
      echo "${dir}/${BINARY_NAME}"
      return 0
    fi
  done
  return 1
}

remove_legacy_symlink() {
  local target="${1}/${LEGACY_SHORT_NAME}"
  if [ -L "$target" ]; then
    rm -f "$target"
    log "removed legacy symlink: $target"
  fi
}

backup_credentials() {
  local backup="${DATA_DIR}.backup.$(date +%Y%m%d%H%M%S)"
  mkdir -p "$backup"
  local f
  for f in $GROUP2_FILES; do
    if [ -e "${DATA_DIR}/${f}" ]; then
      cp "${DATA_DIR}/${f}" "${backup}/${f}"
    fi
  done
  if [ "$(ls -A "$backup" 2>/dev/null)" ]; then
    log "credentials backup: ${backup}"
  else
    rmdir "$backup" 2>/dev/null || true
  fi
}

print_plan() {
  local bin="${BIN_PATH:-}"
  echo "Will remove (Group 1 — binary):"
  [ -n "$bin" ] && echo "  $bin"
  for dir in /usr/local/bin "${HOME}/.local/bin" /opt/homebrew/bin; do
    [ -L "${dir}/${LEGACY_SHORT_NAME}" ] && echo "  ${dir}/${LEGACY_SHORT_NAME}"
  done
  for rc in "${HOME}/.zshrc" "${HOME}/.bashrc"; do
    if [ -f "$rc" ] && grep -qF "$INSTALLER_MARKER" "$rc" 2>/dev/null; then
      echo "  $rc (PATH line)"
    fi
  done

  if [ "$DO_KEEP" = 0 ]; then
    echo ""
    echo "Will consider (Group 2 — credentials):"
    local f
    for f in $GROUP2_FILES; do
      [ -e "${DATA_DIR}/${f}" ] && echo "  ${DATA_DIR}/${f}"
    done
    echo ""
    echo "Will remove (Group 3 — state):"
    for f in $GROUP3_FILES; do
      [ -e "${DATA_DIR}/${f}" ] && echo "  ${DATA_DIR}/${f}"
    done
    for d in $GROUP3_DIRS; do
      [ -e "${DATA_DIR}/${d}" ] && echo "  ${DATA_DIR}/${d}"
    done
  fi
}

cleanup_shell_rc() {
  local rc
  for rc in "${HOME}/.zshrc" "${HOME}/.bashrc"; do
    [ -f "$rc" ] || continue
    if ! grep -qF "$INSTALLER_MARKER" "$rc" 2>/dev/null; then
      continue
    fi
    cp "$rc" "${rc}.opentmd-uninstall.bak"
    awk -v marker="$INSTALLER_MARKER" '
      $0 == marker { skip=1; next }
      skip == 1 && /^export PATH=/ { skip=0; next }
      skip == 1 && /^[[:space:]]*$/ { next }
      { skip=0; print }
    ' "${rc}.opentmd-uninstall.bak" >"$rc"
    log "edited ${rc} (backup: ${rc}.opentmd-uninstall.bak)"
  done
}

prompt_groups() {
  local ans
  DO_G2=0
  DO_G3=1

  if [ "$DO_PURGE" = 1 ]; then
    DO_G2=1
    DO_G3=1
    return 0
  fi
  if [ "$DO_KEEP" = 1 ]; then
    DO_G2=0
    DO_G3=0
    return 0
  fi
  if [ "$DO_YES" = 1 ]; then
    return 0
  fi
  if [ ! -t 0 ]; then
    warn "非交互环境请使用 --yes / --purge / --keep-data / --dry-run"
    exit 2
  fi

  printf "[Group 1] 删除二进制与 PATH 配置? [Y/n]: "
  read -r ans
  case "${ans:-Y}" in [Nn]*) echo "aborted"; exit 1 ;; esac

  printf "[Group 2] 删除 credentials (config.toml, mcp.json)? [y/N]: "
  read -r ans
  case "${ans:-N}" in [Yy]*) DO_G2=1 ;; esac

  printf "[Group 3] 删除 sessions/skills/memory 等状态? [Y/n]: "
  read -r ans
  case "${ans:-Y}" in [Nn]*) DO_G3=0 ;; esac

  printf "继续卸载? [y/N]: "
  read -r ans
  case "${ans:-N}" in [Yy]*) ;; *) echo "aborted"; exit 1 ;; esac
}

remove_group3() {
  local f d
  for f in $GROUP3_FILES; do rm -f "${DATA_DIR}/${f}"; done
  for d in $GROUP3_DIRS; do rm -rf "${DATA_DIR}/${d}"; done
}

remove_group2() {
  backup_credentials
  local f
  for f in $GROUP2_FILES; do rm -f "${DATA_DIR}/${f}"; done
}

main() {
  parse_args "$@"
  log "opentmd-cli uninstall"

  BIN_PATH="$(find_binary || true)"
  print_plan

  [ "$DO_DRY" = 1 ] && exit 0
  prompt_groups

  if [ "${DO_G3:-1}" = 1 ]; then
    remove_group3
    log "removed group 3 state"
  fi
  if [ "${DO_G2:-0}" = 1 ]; then
    remove_group2
    log "removed group 2 credentials"
  fi
  rmdir "$DATA_DIR" 2>/dev/null || true

  cleanup_shell_rc

  if [ -n "${BIN_PATH:-}" ] && [ -e "$BIN_PATH" ]; then
    BIN_DIR="$(dirname "$BIN_PATH")"
    rm -f "$BIN_PATH"
    log "removed: $BIN_PATH"
    remove_legacy_symlink "$BIN_DIR"
  fi
  for dir in /usr/local/bin "${HOME}/.local/bin" /opt/homebrew/bin; do
    remove_legacy_symlink "$dir"
  done

  log "卸载完成"
}

main "$@"
