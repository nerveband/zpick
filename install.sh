#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
INSTALL_DIR="${HOME}/.local/bin"

# ─── Dependency checks ───────────────────────────────────────────────

echo ""
echo "Checking dependencies..."
echo ""

missing=0

if command -v zmosh &>/dev/null; then
  echo "  [ok] zmosh $(zmosh version 2>/dev/null | head -1 | awk '{print $2}')"
else
  echo "  [!!] zmosh not found (REQUIRED)"
  echo "       Install: https://github.com/mmonad/zmosh"
  echo "       brew install mmonad/tap/zmosh"
  missing=1
fi

if command -v zoxide &>/dev/null; then
  echo "  [ok] zoxide $(zoxide --version 2>/dev/null | awk '{print $2}')"
else
  echo "  [--] zoxide not found (optional, for 'z' directory picking)"
  echo "       Install: https://github.com/ajeetdsouza/zoxide"
  echo "       brew install zoxide"
fi

if command -v fzf &>/dev/null; then
  echo "  [ok] fzf $(fzf --version 2>/dev/null | awk '{print $1}')"
else
  echo "  [--] fzf not found (optional, used by zoxide interactive picker)"
  echo "       Install: https://github.com/junegunn/fzf"
  echo "       brew install fzf"
fi

echo ""

if [[ "$missing" -eq 1 ]]; then
  echo "ERROR: Required dependencies missing. Install them first."
  exit 1
fi

# ─── Install script ──────────────────────────────────────────────────

mkdir -p "$INSTALL_DIR"
cp "$SCRIPT_DIR/zmosh-picker" "$INSTALL_DIR/zmosh-picker"
chmod +x "$INSTALL_DIR/zmosh-picker"
echo "Installed zmosh-picker to $INSTALL_DIR/zmosh-picker"
echo ""

# ─── Ask about .zshrc hook ───────────────────────────────────────────

if grep -qF 'zmosh-picker' ~/.zshrc 2>/dev/null; then
  echo "Hook already present in ~/.zshrc"
else
  echo "How do you want to use zmosh-picker?"
  echo ""
  echo "  1) Auto-launch on every new terminal (recommended)"
  echo "     Adds a source hook to ~/.zshrc so the picker shows"
  echo "     every time you open a terminal."
  echo ""
  echo "  2) Manual only"
  echo "     No changes to ~/.zshrc. Run it yourself with:"
  echo "     source ~/.local/bin/zmosh-picker"
  echo ""
  printf "  Choice [1/2]: "
  read -r hook_choice

  case "$hook_choice" in
    2)
      echo ""
      echo "No changes made to ~/.zshrc."
      echo "Run manually with: source ~/.local/bin/zmosh-picker"
      ;;
    *)
      # Default to auto-launch
      if grep -qF 'p10k-instant-prompt' ~/.zshrc 2>/dev/null; then
        p10k_line=$(grep -n 'Enable Powerlevel10k instant prompt' ~/.zshrc | head -1 | cut -d: -f1)
        if [[ -n "$p10k_line" ]]; then
          sed -i '' "${p10k_line}i\\
# zmosh-picker: must run before p10k instant prompt (needs console I/O)\\
[[ -z \"\$ZMX_SESSION\" ]] \\&\\& [[ -f \"\$HOME/.local/bin/zmosh-picker\" ]] \\&\\& source \"\$HOME/.local/bin/zmosh-picker\"
" ~/.zshrc
          echo ""
          echo "Added hook before p10k instant prompt in ~/.zshrc"
        fi
      else
        local_tmp=$(mktemp)
        {
          echo '# zmosh-picker: auto-launch session picker'
          echo '[[ -z "$ZMX_SESSION" ]] && [[ -f "$HOME/.local/bin/zmosh-picker" ]] && source "$HOME/.local/bin/zmosh-picker"'
          echo ''
          cat ~/.zshrc
        } > "$local_tmp"
        mv "$local_tmp" ~/.zshrc
        echo ""
        echo "Added hook to top of ~/.zshrc"
      fi
      ;;
  esac
fi

# ─── Add zpick alias ─────────────────────────────────────────────────

if grep -qF 'alias zpick' ~/.zshrc 2>/dev/null; then
  echo "zpick alias already present in ~/.zshrc"
else
  echo 'alias zpick="ZPICK=1 source $HOME/.local/bin/zmosh-picker"' >> ~/.zshrc
  echo "Added 'zpick' alias to ~/.zshrc"
fi

echo ""
echo "Done! Open a new terminal to try it."
echo "You can also run 'zpick' from any shell to bring up the picker."
echo ""
