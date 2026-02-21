#!/usr/bin/env bash
set -euo pipefail

INSTALL_DIR="${HOME}/.local/bin"

if [[ -f "$INSTALL_DIR/zmosh-picker" ]]; then
  rm "$INSTALL_DIR/zmosh-picker"
  echo "Removed $INSTALL_DIR/zmosh-picker"
fi

if [[ -f ~/.zshrc ]]; then
  sed -i '' '/# zmosh-picker:/d' ~/.zshrc
  sed -i '' '/source.*zmosh-picker/d' ~/.zshrc
  sed -i '' '/_zmosh_picker_once/d' ~/.zshrc
  sed -i '' '/alias zpick/d' ~/.zshrc
  echo "Removed hook and zpick alias from ~/.zshrc"
fi

echo "Done. zmosh-picker has been uninstalled."
