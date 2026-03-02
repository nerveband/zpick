# Changelog

## v3.0.0

- **Remove all legacy migration code** — dropped zmosh-picker hooks, old zpick markers, and guard→hook block rename migration. Fresh install only; no backward compatibility shims.
- **Remove TERM-fix code** — Ghostty terminal detection and `export TERM=xterm-ghostty` injection removed from both bash/zsh and fish hooks.
- **Consolidate config directory** — update cache moved from `~/.zpick/` to `~/.config/zpick/`, matching all other config files.
- **Deduplicate utilities** — shared `DetectShell()` in `backend` package replaces copies in `hook` and `check`; `guard.ConfigPath()` reuses `backend.ConfigDir()`.
- **Single-read `HookStatus()`** — reads shell config once to check both hook and guard presence, replacing triple file reads during upgrade.
- **Shared install summary** — `printInstallSummary()` used by both shell and fish installers.

## v2.9.3

- **Fix hook refresh on upgrade** — only refresh the shell hook when `zp upgrade` actually downloads a new version.

## v2.9.2

- **Auto-install hook** — `make install` now installs the shell hook automatically. `zp upgrade` auto-updates the hook after downloading a new version.

## v2.9.1

- **Hook is required** — `zp check` and first-run messaging clarify that the shell hook is required for zp to work. Auto-sudo symlink creation. Block markers renamed from "guard" to "hook".

## v2.9.0

- **Separate hook and guard commands** — `install-hook` and `install-guard` are now distinct commands. Guard wrappers are opt-in via `zp install-guard` instead of `zp install-hook --guard`. New `remove-hook` and `remove-guard` commands for targeted removal.
- **Guard explanation** — `zp guard` with no arguments prints a clear explanation of what guard does, how it works, and its limitations.
- **Deprecated flags** — `--guard` and `--remove` flags on `install-hook` still work but print deprecation notices pointing to the new commands.

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
