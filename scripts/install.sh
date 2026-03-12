#!/usr/bin/env sh
set -eu

VERSION="${1:-latest}"
REPO="example/ollamon"
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
esac

if [ "$VERSION" = "latest" ]; then
  URL="https://github.com/${REPO}/releases/latest/download/ollamon_latest_${OS}_${ARCH}.tar.gz"
else
  URL="https://github.com/${REPO}/releases/download/${VERSION}/ollamon_${VERSION#v}_${OS}_${ARCH}.tar.gz"
fi

echo "Download URL örneği: $URL"
echo "Release ad formatını repo adına göre eşleştirmen gerekecek."