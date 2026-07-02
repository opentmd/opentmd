#!/usr/bin/env bash
# OpenTMD npm package publish script
#
# Usage:
#   ./packages/npm/scripts/build.sh 0.1.0
#   ./packages/npm/scripts/build.sh 0.1.0 --dry-run
#
# Requires: curl, npm, node

set -euo pipefail

VERSION="${1:-}"
shift || true
NPM_EXTRA="${*:-}"
TAG_VERSION="${VERSION#v}"

if [ -z "$VERSION" ]; then
  echo "Usage: $0 <version> [npm publish flags]" >&2
  exit 1
fi

NPM_DIR="$(cd "$(dirname "$0")/.." && pwd)"
REPO="${OPENTMD_REPO:-opentmd/opentmd-cli}"
RELEASE_BASE="https://github.com/${REPO}/releases/download/v${TAG_VERSION}"
WORK_DIR="$(mktemp -d)"
trap 'rm -rf "$WORK_DIR"' EXIT

if command -v curl >/dev/null 2>&1; then
  DOWNLOAD="curl -fsSL --connect-timeout 30 --retry 3 -o"
else
  echo "curl is required" >&2
  exit 1
fi

PLATFORMS=(
  "darwin-arm64:darwin:arm64:arm64"
  "darwin-x64:darwin:amd64:x64"
  "linux-arm64:linux:arm64:arm64"
  "linux-x64:linux:amd64:x64"
)

publish_platform() {
  local tag="$1" os="$2" arch="$3" npm_cpu="$4"
  local dir="${WORK_DIR}/${tag}"
  mkdir -p "${dir}/bin"

  cat >"${dir}/package.json" <<EOF
{
  "name": "@opentmd/cli-${tag}",
  "version": "${VERSION}",
  "os": ["${os}"],
  "cpu": ["${npm_cpu}"],
  "files": ["bin/"]
}
EOF

  local urls=(
    "${RELEASE_BASE}/opentmd_${os}_${arch}.tar.gz"
    "${RELEASE_BASE}/opentmd_${VERSION}_${os}_${arch}.tar.gz"
    "${RELEASE_BASE}/opentmd_${os}_${arch}"
  )
  local url dest="${dir}/artifact" bin="${dir}/bin/opentmd" ok=0

  for url in "${urls[@]}"; do
    echo "  ↓ ${tag}: ${url}"
    if $DOWNLOAD "$dest" "$url" 2>/dev/null; then
      if file "$dest" 2>/dev/null | grep -qi 'gzip compressed'; then
        tar -xzf "$dest" -C "$dir"
        rm -f "$dest"
        if [ -f "${dir}/opentmd" ]; then
          mv "${dir}/opentmd" "$bin"
          ok=1
          break
        fi
      else
        mv "$dest" "$bin"
        ok=1
        break
      fi
    fi
    rm -f "$dest"
  done

  if [ "$ok" -ne 1 ]; then
    echo "  ⚠ binary not found for ${tag}, skipping"
    return 0
  fi

  chmod +x "$bin"
  (
    cd "$dir"
    npm publish --access public $NPM_EXTRA
  )
  echo "  ✓ @opentmd/cli-${tag}@${VERSION}"
}

echo ""
echo "Publishing @opentmd/cli v${VERSION}"
echo ""

for entry in "${PLATFORMS[@]}"; do
  IFS=: read -r tag os arch npm_cpu <<<"$entry"
  publish_platform "$tag" "$os" "$arch" "$npm_cpu"
done

CORE_DIR="${WORK_DIR}/core"
mkdir -p "${CORE_DIR}/bin"
cp "${NPM_DIR}/package.json" "${CORE_DIR}/"
cp "${NPM_DIR}/bin/opentmd.js" "${CORE_DIR}/bin/"

node -e "
const fs = require('fs');
const pkg = JSON.parse(fs.readFileSync('${CORE_DIR}/package.json', 'utf8'));
pkg.version = '${VERSION}';
pkg.optionalDependencies = {
  '@opentmd/cli-darwin-arm64': '${VERSION}',
  '@opentmd/cli-darwin-x64': '${VERSION}',
  '@opentmd/cli-linux-arm64': '${VERSION}',
  '@opentmd/cli-linux-x64': '${VERSION}'
};
fs.writeFileSync('${CORE_DIR}/package.json', JSON.stringify(pkg, null, 2) + '\n');
"

(
  cd "$CORE_DIR"
  npm publish --access public $NPM_EXTRA
)
echo "  ✓ @opentmd/cli@${VERSION} (core)"
echo ""
echo "All done!"
