# zpick — Agent Instructions

## Releases

- **Always use goreleaser** for releases so pre-built binaries and `checksums.txt` are attached to the GitHub release. The built-in `zp upgrade` command depends on these assets (`go-selfupdate` with `ChecksumValidator`).
- Never create releases with just `gh release create` — that only produces source archives, which breaks the self-updater.
- Release flow: `git tag vX.Y.Z && git push origin vX.Y.Z && goreleaser release --clean`
- If `GITHUB_TOKEN` is not set, warn the user before proceeding. Do not create a release without binaries.
- goreleaser ships two binaries: `zp` (primary) and `zpick` (deprecation shim).

## Build & Install

- `make build` — build `zp` binary with version from git tags
- `make install` — build, sign, and copy to `~/.local/bin/`
- `make test` — run all tests
- Binary name is `zp` (renamed from `zpick` in v2.4.0). The `zpick` binary is a transitional shim.

## macOS Code Signing (CRITICAL)

- **Always ad-hoc sign the binary on macOS** after building. Without signing, macOS silently kills the binary (SIGKILL, no error message).
- `make install` handles this automatically — it copies the binary, strips xattrs (`xattr -cr`), then signs (`codesign -fs -`).
- The repo lives inside Dropbox, so `cp` carries over `com.dropbox.attrs` and `com.apple.provenance` xattrs. These must be stripped **before** signing, and signing must happen **after** copying to the install path.
- If a user reports "zp doesn't load" or "zp just exits silently", check signing first: `codesign -v ~/.local/bin/zp`

## Project Notes

- This is a **multi-backend** session picker supporting zmosh, zmx, tmux, and shpool via the `backend.Backend` interface.
- Backend is configured in `~/.config/zpick/backend`. Auto-detected if only one backend is available.
- The picker suppresses backend command output and displays its own clean status messages.
- The TUI renders to `/dev/tty` and outputs only the final shell command to stdout, used via `eval "$(zp)"`.
- Shell hooks support zsh, bash, and fish. Fish hooks go to `~/.config/fish/conf.d/zp.fish`.
- Hook env var guard checks all backend session vars (ZMX_SESSION, TMUX, SHPOOL_SESSION_NAME).
