#!/usr/bin/env bash
#
# render-og-image.sh — regenerate static/og-image.png from scripts/og-image.svg.
# Invoked by `make og-image`.
#
# Renders with rsvg-convert (librsvg; `brew install librsvg`). The text is set
# in the site's faces, IBM Plex Sans and JetBrains Mono, as TTFs downloaded
# once into .cache/fonts/ (gitignored) from a pinned google/fonts commit — the
# woff2 files fetch-fonts.sh pulls for the browser are not readable by
# fontconfig. A private fontconfig sees only that directory, and
# PANGOCAIRO_BACKEND forces pango onto fontconfig (on macOS it defaults to
# CoreText, which ignores FONTCONFIG_FILE and substitutes a system sans).
# Same librsvg, same fonts → byte-identical PNG; `--check` verifies that.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SRC="$ROOT/scripts/og-image.svg"
OUT="$ROOT/static/og-image.png"
FONTS="$ROOT/.cache/fonts"
GF="https://raw.githubusercontent.com/google/fonts/5bb28733cec313b63fbd415b574b5ae35c8aa8b8/ofl"

command -v rsvg-convert >/dev/null || { echo "rsvg-convert not found (brew install librsvg)" >&2; exit 1; }

mkdir -p "$FONTS"
fetch() { [ -f "$FONTS/$2" ] || { echo "fetching $2"; curl -fsSL -o "$FONTS/$2" "$GF/$1"; }; }
fetch 'ibmplexsans/IBMPlexSans%5Bwdth%2Cwght%5D.ttf' IBMPlexSans.ttf
fetch 'jetbrainsmono/JetBrainsMono%5Bwght%5D.ttf'    JetBrainsMono.ttf

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT
cat > "$TMP/fonts.conf" <<CONF
<?xml version="1.0"?>
<!DOCTYPE fontconfig SYSTEM "fonts.dtd">
<fontconfig>
  <dir>$FONTS</dir>
  <cachedir>$TMP/cache</cachedir>
</fontconfig>
CONF

TARGET="$OUT"
[ "${1:-}" = "--check" ] && TARGET="$TMP/og-image.png"

PANGOCAIRO_BACKEND=fontconfig FONTCONFIG_FILE="$TMP/fonts.conf" rsvg-convert \
  --width 1200 --height 630 --format png --output "$TARGET" "$SRC"

if [ "${1:-}" = "--check" ]; then
  cmp -s "$TARGET" "$OUT" && echo "ok: $OUT matches $SRC" || { echo "DIFFERS: run make og-image" >&2; exit 1; }
else
  echo "wrote $OUT ($(wc -c < "$OUT" | tr -d ' ') bytes)"
fi
