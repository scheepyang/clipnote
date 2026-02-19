#!/usr/bin/env bash
set -euo pipefail

REPO="nevertomica/clipnote"
BIN_DIR="${CLAUDE_PLUGIN_ROOT:-$(dirname "$(dirname "$0")")}/bin"
BIN_PATH="${BIN_DIR}/clipnote"

# Check tmux dependency
if ! command -v tmux &>/dev/null; then
  echo "clipnote: tmux is required but not installed." >&2
  echo "  macOS:  brew install tmux" >&2
  echo "  Linux:  sudo apt install tmux" >&2
  exit 1
fi

# Skip if binary already exists
if [[ -x "$BIN_PATH" ]]; then
  exit 0
fi

# Detect platform
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  arm64)   ARCH="arm64" ;;
  *)
    echo "clipnote: unsupported architecture: $ARCH" >&2
    exit 1
    ;;
esac

# Get latest release tag
TAG="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": *"//;s/".*//')"
if [[ -z "$TAG" ]]; then
  echo "clipnote: failed to fetch latest release tag" >&2
  exit 1
fi

ASSET="clipnote-${OS}-${ARCH}"
URL="https://github.com/${REPO}/releases/download/${TAG}/${ASSET}"

echo "clipnote: downloading ${ASSET} (${TAG})..."
mkdir -p "$BIN_DIR"
curl -fsSL -o "$BIN_PATH" "$URL"
chmod +x "$BIN_PATH"
echo "clipnote: installed to ${BIN_PATH}"
