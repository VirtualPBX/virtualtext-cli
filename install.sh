#!/usr/bin/env bash
# Usage: curl -fsSL https://raw.githubusercontent.com/VirtualPBX/virtualtext-cli/main/install.sh | bash
set -euo pipefail

REPO="${VT_REPO:-VirtualPBX/virtualtext-cli}"
BINARY="vt"
INSTALL_DIR="${VT_BIN_DIR:-}"

if [ -z "$INSTALL_DIR" ]; then
  if [ -d "$HOME/.local/bin" ]; then
    INSTALL_DIR="$HOME/.local/bin"
  else
    INSTALL_DIR="/usr/local/bin"
  fi
fi

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac
case "$OS" in
  darwin|linux) ;;
  *) echo "Unsupported OS: $OS" >&2; exit 1 ;;
esac

install_from_release() {
  local latest archive url checksum_url tmp expected actual
  latest="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | head -1 | sed -E 's/.*"v?([^"]+)".*/\1/')"
  [ -n "$latest" ] || return 1
  archive="${BINARY}_${latest}_${OS}_${ARCH}.tar.gz"
  url="https://github.com/${REPO}/releases/download/v${latest}/${archive}"
  checksum_url="https://github.com/${REPO}/releases/download/v${latest}/checksums.txt"
  tmp="$(mktemp -d)"
  trap 'rm -rf "$tmp"' RETURN
  echo "Downloading ${BINARY} v${latest} for ${OS}/${ARCH}..."
  curl -fsSL "$url" -o "$tmp/$archive"
  curl -fsSL "$checksum_url" -o "$tmp/checksums.txt"
  expected="$(grep -F -- "$archive" "$tmp/checksums.txt" | awk '{print $1}')"
  if [ -n "$expected" ]; then
    if command -v sha256sum >/dev/null 2>&1; then
      actual="$(sha256sum "$tmp/$archive" | awk '{print $1}')"
    else
      actual="$(shasum -a 256 "$tmp/$archive" | awk '{print $1}')"
    fi
    if [ "$actual" != "$expected" ]; then
      echo "Checksum mismatch" >&2
      return 1
    fi
  fi
  tar -xzf "$tmp/$archive" -C "$tmp"
  mkdir -p "$INSTALL_DIR"
  if [ -w "$INSTALL_DIR" ]; then
    install "$tmp/$BINARY" "$INSTALL_DIR/$BINARY"
  else
    sudo install "$tmp/$BINARY" "$INSTALL_DIR/$BINARY"
  fi
  echo "Installed $INSTALL_DIR/$BINARY"
}

if ! install_from_release; then
  echo "No GitHub release found; building with go install..."
  if ! command -v go >/dev/null 2>&1; then
    echo "Install Go, or wait for a GitHub release of ${REPO}." >&2
    exit 1
  fi
  GOBIN="$INSTALL_DIR" go install "github.com/${REPO}/cmd/vt@latest"
  echo "Installed $INSTALL_DIR/$BINARY"
fi

echo "Next: vt auth login --host https://YOUR_INSTANCE.virtualtext.app"
