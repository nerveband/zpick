# zpick — Agent Instructions

## Releases

- **Always use goreleaser** for releases so pre-built binaries and `checksums.txt` are attached to the GitHub release. The built-in `zp upgrade` command depends on these assets (`go-selfupdate` with `ChecksumValidator`).
- Never create releases with just `gh release create` — that only produces source archives, which breaks the self-updater.
- Release flow: `git tag vX.Y.Z && git push origin vX.Y.Z && goreleaser release --clean`
- If `GITHUB_TOKEN` is not set, warn the user before proceeding. Do not create a release without binaries.
- goreleaser ships two binaries: `zp` (primary) and `zpick` (deprecation shim).

## Build & Install

- `make build` — build `zp` binary with version from git tags
- `make install` — build and copy to `~/.local/bin/`
- `make test` — run all tests
- Binary name is `zp` (renamed from `zpick` in v2.4.0). The `zpick` binary is a transitional shim.

## Project Notes

- This is a **multi-backend** session picker supporting zmosh, zmx, tmux, and shpool via the `backend.Backend` interface.
- Backend is configured in `~/.config/zpick/backend`. Auto-detected if only one backend is available.
- The picker suppresses backend command output and displays its own clean status messages.
- The TUI renders to `/dev/tty` and outputs only the final shell command to stdout, used via `eval "$(zp)"`.
- Shell hooks support zsh, bash, and fish. Fish hooks go to `~/.config/fish/conf.d/zp.fish`.
- Hook env var guard checks all backend session vars (ZMX_SESSION, TMUX, SHPOOL_SESSION_NAME).
