#!/usr/bin/env sh
set -eu

REPO="${OLLAMON_REPO:-hbasria/ollamon}"
VERSION="${OLLAMON_VERSION:-latest}"
RUN_AFTER_INSTALL=1

usage() {
  cat <<'EOF'
Usage: install.sh [--version <tag>] [--install-dir <path>] [--no-run]

Options:
  --version, -v     Release tag (default: latest)
  --install-dir     Target install directory
  --no-run          Install only, do not start ollamon
  --help, -h        Show this help

Environment:
  OLLAMON_REPO        GitHub repository (default: hbasria/ollamon)
  OLLAMON_VERSION     Version override (same as --version)
  OLLAMON_INSTALL_DIR Install directory override

Supported targets:
  darwin/amd64
  darwin/arm64 (Apple Silicon)
  linux/amd64
  linux/arm64

One-liner:
  curl -fsSL https://raw.githubusercontent.com/hbasria/ollamon/main/scripts/install.sh | sh
EOF
}

INSTALL_DIR="${OLLAMON_INSTALL_DIR:-}"

while [ "$#" -gt 0 ]; do
  case "$1" in
    --version|-v)
      VERSION="${2:-}"
      shift 2
      ;;
    --install-dir)
      INSTALL_DIR="${2:-}"
      shift 2
      ;;
    --no-run)
      RUN_AFTER_INSTALL=0
      shift
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      VERSION="$1"
      shift
      ;;
  esac
done

if [ -z "$VERSION" ]; then
  echo "Version empty. Use --version <tag> or omit for latest." >&2
  exit 1
fi

if ! command -v curl >/dev/null 2>&1; then
  echo "curl not found. Please install curl and retry." >&2
  exit 1
fi

if ! command -v tar >/dev/null 2>&1; then
  echo "tar not found. Please install tar and retry." >&2
  exit 1
fi

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: $ARCH" >&2
    echo "Supported: darwin/amd64, darwin/arm64, linux/amd64, linux/arm64" >&2
    exit 1
    ;;
esac

case "$OS/$ARCH" in
  darwin/amd64|darwin/arm64|linux/amd64|linux/arm64)
    ;;
  *)
    echo "Unsupported target: $OS/$ARCH" >&2
    echo "Supported: darwin/amd64, darwin/arm64, linux/amd64, linux/arm64" >&2
    exit 1
    ;;
esac

if [ -z "$INSTALL_DIR" ]; then
  if [ -d "/usr/local/bin" ] && [ -w "/usr/local/bin" ]; then
    INSTALL_DIR="/usr/local/bin"
  else
    INSTALL_DIR="$HOME/.local/bin"
  fi
fi

if [ "$VERSION" = "latest" ]; then
  TAG="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -n 1)"
  if [ -z "$TAG" ]; then
    echo "Could not resolve latest release for ${REPO}" >&2
    exit 1
  fi
else
  TAG="$VERSION"
fi

VERSION_NO_V="${TAG#v}"
ASSET="ollamon_${VERSION_NO_V}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${TAG}/${ASSET}"

TMP_DIR="$(mktemp -d)"
cleanup() {
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT INT TERM

ARCHIVE_PATH="$TMP_DIR/$ASSET"

echo "Downloading: $URL"
curl -fL "$URL" -o "$ARCHIVE_PATH"

tar -xzf "$ARCHIVE_PATH" -C "$TMP_DIR"

if [ ! -f "$TMP_DIR/ollamon" ]; then
  echo "Archive does not contain 'ollamon' binary." >&2
  exit 1
fi

mkdir -p "$INSTALL_DIR"
TARGET="$INSTALL_DIR/ollamon"
cp "$TMP_DIR/ollamon" "$TARGET"
chmod +x "$TARGET"

echo "Installed: $TARGET"

if ! command -v ollamon >/dev/null 2>&1; then
  echo "Note: '$INSTALL_DIR' may not be in PATH."
  echo "Run with: $TARGET"
fi

if [ "$RUN_AFTER_INSTALL" -eq 1 ]; then
  echo "Starting ollamon..."
  "$TARGET"
fi