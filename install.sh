#!/usr/bin/env bash
set -euo pipefail

# zpick installer
# Downloads pre-built binary from GitHub releases, or builds from source.

REPO="nerveband/zpick"
INSTALL_DIR="${HOME}/.local/bin"

echo ""
echo "Installing zpick..."
echo ""

# ─── Detect platform ─────────────────────────────────────────────────

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  arm64)   ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

echo "  Platform: ${OS}/${ARCH}"

# ─── Try downloading pre-built binary ────────────────────────────────

mkdir -p "$INSTALL_DIR"

if command -v curl &>/dev/null; then
  LATEST_URL="https://api.github.com/repos/${REPO}/releases/latest"
  TAG=$(curl -fsSL "$LATEST_URL" 2>/dev/null | grep '"tag_name"' | head -1 | cut -d'"' -f4 || true)

  if [[ -n "$TAG" ]]; then
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${TAG}/zpick_${TAG#v}_${OS}_${ARCH}.tar.gz"
    echo "  Downloading ${TAG}..."

    if curl -fsSL "$DOWNLOAD_URL" | tar xz -C "$INSTALL_DIR" zpick 2>/dev/null; then
      chmod +x "$INSTALL_DIR/zpick"
      echo "  Installed zpick ${TAG} to ${INSTALL_DIR}/zpick"
      "$INSTALL_DIR/zpick" install-hook
      echo ""
      echo "Done! Open a new terminal to try it."
      exit 0
    fi
    echo "  Download failed, falling back to go install..."
  fi
fi

# ─── Fallback: go install ────────────────────────────────────────────

if command -v go &>/dev/null; then
  echo "  Building from source..."
  go install "github.com/${REPO}/cmd/zpick@latest"
  echo "  Installed via go install"

  # Ensure GOBIN is in PATH
  GOBIN="$(go env GOPATH)/bin"
  if [[ -f "$GOBIN/zpick" ]]; then
    "$GOBIN/zpick" install-hook
  fi
  echo ""
  echo "Done! Open a new terminal to try it."
  exit 0
fi

# ─── Fallback: build from local source ───────────────────────────────

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
if [[ -f "$SCRIPT_DIR/go.mod" ]]; then
  echo "  Building from local source..."
  cd "$SCRIPT_DIR"
  go build -o "$INSTALL_DIR/zpick" ./cmd/zpick
  chmod +x "$INSTALL_DIR/zpick"
  "$INSTALL_DIR/zpick" install-hook
  echo ""
  echo "Done! Open a new terminal to try it."
  exit 0
fi

echo "ERROR: Could not install. Install Go or download a release from:"
echo "  https://github.com/${REPO}/releases"
exit 1
