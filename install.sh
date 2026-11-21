#!/bin/bash
set -euo pipefail

REPO="toms74209200/shampoo"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"

OS=$(uname -s)
ARCH=$(uname -m)

case "$OS/$ARCH" in
  Linux/x86_64) ARTIFACT="shampoo-linux-x86_64" ;;
  Linux/aarch64 | Linux/arm64) ARTIFACT="shampoo-linux-aarch64" ;;
  Darwin/arm64) ARTIFACT="shampoo-macos-aarch64" ;;
  *)
    echo "Unsupported platform: $OS $ARCH" >&2
    exit 1
    ;;
esac

TAG=$(curl -sf "https://api.github.com/repos/$REPO/releases/latest" \
  | grep -o '"tag_name": "[^"]*"' \
  | cut -d'"' -f4)

if [ -z "$TAG" ]; then
  echo "Failed to fetch latest release tag" >&2
  exit 1
fi

URL="https://github.com/$REPO/releases/download/$TAG/$ARTIFACT"

mkdir -p "$INSTALL_DIR"
echo "Downloading $ARTIFACT ($TAG)..."
curl -fsSL "$URL" -o "$INSTALL_DIR/shampoo"
chmod +x "$INSTALL_DIR/shampoo"
echo "Installed shampoo $TAG to $INSTALL_DIR/shampoo"
