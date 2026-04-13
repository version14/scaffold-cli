#!/bin/sh
# Install dot — universal project companion
# Usage: curl -fsSL https://raw.githubusercontent.com/version14/dot/main/install.sh | sh
set -e

REPO="version14/dot"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
BINARY="dot"

# Detect OS and architecture.
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: $ARCH" >&2
    exit 1
    ;;
esac

# Fetch the latest release tag.
VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
  | grep '"tag_name"' \
  | sed 's/.*"v\([^"]*\)".*/\1/')

if [ -z "$VERSION" ]; then
  echo "Could not determine latest version. Check your internet connection." >&2
  exit 1
fi

echo "Installing dot v${VERSION} (${OS}/${ARCH}) → ${INSTALL_DIR}/${BINARY}"

# Build the download URL.
EXT="tar.gz"
[ "$OS" = "windows" ] && EXT="zip"
URL="https://github.com/${REPO}/releases/download/v${VERSION}/dot_${VERSION}_${OS}_${ARCH}.${EXT}"

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

# Download and extract.
curl -fsSL "$URL" -o "$TMP/archive.${EXT}"

if [ "$EXT" = "zip" ]; then
  unzip -q "$TMP/archive.zip" -d "$TMP"
else
  tar -xzf "$TMP/archive.tar.gz" -C "$TMP"
fi

# Install the binary.
if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP/dot" "$INSTALL_DIR/$BINARY"
else
  sudo mv "$TMP/dot" "$INSTALL_DIR/$BINARY"
fi
chmod +x "$INSTALL_DIR/$BINARY"

echo "✓  dot v${VERSION} installed to ${INSTALL_DIR}/${BINARY}"
echo "   Run 'dot version' to verify."
