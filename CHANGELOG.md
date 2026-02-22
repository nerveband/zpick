# Changelog

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
