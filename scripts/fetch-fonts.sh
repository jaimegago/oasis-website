#!/usr/bin/env bash
# fetch-fonts.sh — Download self-hosted web fonts for the OASIS site.
# Idempotent: skips downloads if files already exist.
# Fonts are gitignored; this script runs as part of make serve / make build.

set -euo pipefail

FONT_DIR="$(cd "$(dirname "$0")/.." && pwd)/static/fonts"
mkdir -p "$FONT_DIR"

download() {
  local url="$1"
  local dest="$2"
  if [ -f "$dest" ]; then
    return 0
  fi
  echo "Downloading $(basename "$dest")..."
  curl -fsSL -o "$dest" "$url"
}

# IBM Plex Sans — 400, 500, 600, 700 (latin)
# URLs from Google Fonts CSS API (woff2, latin subset)
download "https://fonts.gstatic.com/s/ibmplexsans/v19/zYXgKVElMYYaJe8bpLHnCwDKhdHeFaxOedc.woff2" \
  "$FONT_DIR/IBMPlexSans-Regular.woff2"

download "https://fonts.gstatic.com/s/ibmplexsans/v19/zYX9KVElMYYaJe8bpLHnCwDKjSL9AIFsdP3pBms.woff2" \
  "$FONT_DIR/IBMPlexSans-Medium.woff2"

download "https://fonts.gstatic.com/s/ibmplexsans/v19/zYX9KVElMYYaJe8bpLHnCwDKjXr8AIFsdP3pBms.woff2" \
  "$FONT_DIR/IBMPlexSans-SemiBold.woff2"

download "https://fonts.gstatic.com/s/ibmplexsans/v19/zYX9KVElMYYaJe8bpLHnCwDKjWr7AIFsdP3pBms.woff2" \
  "$FONT_DIR/IBMPlexSans-Bold.woff2"

# JetBrains Mono — 400 (latin)
download "https://fonts.gstatic.com/s/jetbrainsmono/v18/tDbY2o-flEEny0FZhsfKu5WU4zr3E_BX0PnT8RD8yKxTOlOV.woff2" \
  "$FONT_DIR/JetBrainsMono-Regular.woff2"

echo "Fonts ready in $FONT_DIR"
