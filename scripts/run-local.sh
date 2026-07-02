#!/usr/bin/env bash
# 仓库内快捷启动：自动 build 缺失的二进制后 exec
#
# 用法:
#   ./opentmd [args...]

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN="${ROOT}/bin/opentmd"
if [ ! -x "$BIN" ]; then
  "${ROOT}/scripts/build.sh"
fi

exec "$BIN" "$@"
