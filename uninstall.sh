#!/bin/sh
# Uninstall dot — universal project companion
# Usage: curl -fsSL https://raw.githubusercontent.com/version14/dot/main/uninstall.sh | sh
set -e

BINARY="dot"

# Resolve where the binary currently lives.
if command -v "$BINARY" > /dev/null 2>&1; then
  INSTALL_PATH=$(command -v "$BINARY")
else
  # Fall back to the default install location if dot is not on PATH.
  INSTALL_PATH="${INSTALL_DIR:-/usr/local/bin}/$BINARY"
fi

if [ ! -f "$INSTALL_PATH" ]; then
  echo "dot is not installed at $INSTALL_PATH — nothing to do."
  exit 0
fi

echo "Removing $INSTALL_PATH"

if [ -w "$(dirname "$INSTALL_PATH")" ]; then
  rm -f "$INSTALL_PATH"
else
  sudo rm -f "$INSTALL_PATH"
fi

echo "✓  dot uninstalled."
echo "   Project .dot/ directories are untouched — remove them manually if needed."
