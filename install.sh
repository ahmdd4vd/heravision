#!/bin/sh
# HeraVision one-line installer (macOS / Linux)
# usage: curl -fsSL https://raw.githubusercontent.com/ahmdd4vd/heravision/main/install.sh | bash
set -e

REPO="ahmdd4vd/heravision"
case "$(uname -s)" in
  Linux) OSN=Linux ;;
  Darwin) OSN=Darwin ;;
  *) echo "unsupported os: $(uname -s)" >&2; exit 1 ;;
esac
case "$(uname -m)" in
  x86_64|amd64) ARN=x86_64 ;;
  arm64|aarch64) ARN=arm64 ;;
  *) echo "unsupported arch: $(uname -m)" >&2; exit 1 ;;
esac

echo "resolving latest release for ${OSN} ${ARN}..."
ASSETS=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest")
URL=$(printf "%s" "$ASSETS" | grep -o "https://[^\"']*heravision_[^\"']*_${OSN}_${ARN}\.tar\.gz" | head -n 1)
if [ -z "$URL" ]; then
  echo "no release asset found for ${OSN} ${ARN}" >&2
  exit 1
fi

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT
echo "downloading $(basename "$URL")"
curl -fsSL "$URL" | tar -xz -C "$TMP"
if [ ! -f "$TMP/heravision" ]; then
  echo "archive did not contain the heravision binary" >&2
  exit 1
fi

DEST="/usr/local/bin"
if [ ! -w "$DEST" ]; then
  DEST="$HOME/.local/bin"
  mkdir -p "$DEST"
fi
install -m 0755 "$TMP/heravision" "$DEST/heravision"
case ":$PATH:" in
  *":$DEST:"*) ;;
  *) echo "note: add $DEST to your PATH (export PATH=\"$DEST:\$PATH\")" ;;
esac

echo "installed: $DEST/heravision"
# interactive setup (OCR + agent config); /dev/tty keeps prompts alive under curl | bash
if [ -r /dev/tty ]; then
  "$DEST/heravision" setup < /dev/tty || true
else
  echo "run: heravision setup   # enable OCR + connect your AI agent"
fi
