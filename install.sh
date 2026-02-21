#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
INSTALL_DIR="${HOME}/.local/bin"
HOOK='[[ -z "$ZMX_SESSION" ]] && command -v zmosh-picker &>/dev/null && zmosh-picker'

mkdir -p "$INSTALL_DIR"
cp "$SCRIPT_DIR/zmosh-picker" "$INSTALL_DIR/zmosh-picker"
chmod +x "$INSTALL_DIR/zmosh-picker"
echo "Installed zmosh-picker to $INSTALL_DIR/zmosh-picker"

if ! grep -qF 'zmosh-picker' ~/.zshrc 2>/dev/null; then
  echo "" >> ~/.zshrc
  echo "# zmosh-picker: auto-launch session picker" >> ~/.zshrc
  echo "$HOOK" >> ~/.zshrc
  echo "Added hook to ~/.zshrc"
else
  echo "Hook already present in ~/.zshrc"
fi

echo "Done. Open a new terminal to try it."
