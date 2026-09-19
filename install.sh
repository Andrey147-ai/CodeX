#!/bin/sh
# Install CodeX in one command:
#   curl -fsSL https://raw.githubusercontent.com/Andrey147-ai/CodeX/main/install.sh | sh
# Options: VERSION=v0.23.0 DIR=~/.codex/bin sh install.sh
set -eu
REPO="Andrey147-ai/CodeX"
VERSION="${VERSION:-latest}"
DIR="${DIR:-$HOME/.codex/bin}"
mkdir -p "$DIR"

if [ "$VERSION" = "latest" ] || [ -z "$VERSION" ]; then
  API="https://api.github.com/repos/$REPO/releases/latest"
else
  API="https://api.github.com/repos/$REPO/releases/tags/$VERSION"
fi
URL=$(curl -fsSL "$API" | grep -o '"browser_download_url": *"[^"]*codex-linux[^"]*"' | head -1 | cut -d'"' -f4)
if [ -z "$URL" ]; then
  echo "codex-linux not found in release. Push a tag (release.yml publishes it)." >&2
  exit 1
fi
echo "Downloading CodeX ($VERSION) ..."
curl -fsSL -o "$DIR/codex" "$URL"
chmod +x "$DIR/codex"
"$DIR/codex" version
case ":$PATH:" in
  *":$DIR:"*) ;;
  *) echo "Add to PATH: export PATH=\"$DIR:\$PATH\"" ;;
esac
echo "OK: $DIR/codex"
