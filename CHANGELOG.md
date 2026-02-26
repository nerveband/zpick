# Changelog

## v2.8.0

- **Key mode toggle** — switch session labels between numbers-first (`1-9,a-y`) and letters-first (`a-y,1-9`). Useful on mobile keyboards where letters are the default view. Press `h` then `l` to toggle. Saved to `~/.config/zpick/keys`.

## v2.7.0

- **In-session switching** — run `zp` inside an active session to switch, create, or kill sessions without remembering backend-specific detach commands. Works with all backends (tmux, zellij, zmosh, zmx, shpool).
- **Current session marker** — when inside a session, the picker header shows `(in: name ←)` and marks your current session with `←` in the session list.
- **System PATH symlink** — `zp install-hook` creates a `/usr/local/bin/zp` symlink so `mosh host -- zp` works out of the box. Prints a `sudo` hint if permissions require it.
- **Guard block is optional** — the shell hook no longer includes guard wrappers by default. Only added if `~/.config/zpick/guard.conf` exists with apps listed.
- **`zp resume` subcommand** — internal command used by shell hooks to handle switch-target handoff after detaching.

## v2.0.0

- **Renamed to zpick** — binary, module, config dir all renamed from `zmosh-picker` to `zpick`
- **Eval refactor** — TUI renders to `/dev/tty`, only the final shell command goes to stdout. Hook becomes `eval "$(zpick)"`. Fixes shell lifecycle: when zmosh session ends, the terminal closes properly instead of dropping to an outer shell.
- **Script-friendly** — `zpick install-hook` auto-detects your shell (zsh/bash) and adds the hook. `zpick | cat` outputs only the command string.

## v1.1.1

- **Cleaner kill output** — suppress noisy zmosh output during kill; only show the picker's own status message

## v1.1.0

- **Kill mode returns to menu** — after killing a session (or cancelling), you stay in the picker with a refreshed session list instead of dropping back to the shell
- **Cleaner default names** — new sessions use the bare directory name (e.g. `myproject`) instead of always appending `-1`; only adds `-2`, `-3`, etc. when a conflict exists
- **Custom name flow** — press `c` to type any session name, then choose where to create it (`enter` for current dir, `z` for zoxide picker, `esc` to cancel)

## v1.0.0

- Initial release: single-keypress session picker for zmosh
- Kill mode (`k`), zoxide integration (`z`), date suffix (`d`)
