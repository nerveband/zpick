# zpick — Agent Instructions

## Releases

- **Always use goreleaser** for releases so pre-built binaries and `checksums.txt` are attached to the GitHub release. The built-in `zpick upgrade` command depends on these assets (`go-selfupdate` with `ChecksumValidator`).
- Never create releases with just `gh release create` — that only produces source archives, which breaks the self-updater.
- Release flow: `git tag vX.Y.Z && git push origin vX.Y.Z && goreleaser release --clean`
- If `GITHUB_TOKEN` is not set, warn the user before proceeding. Do not create a release without binaries.

## Build & Install

- `make build` — build with version from git tags
- `make install` — build and copy to `~/.local/bin/`
- `make test` — run all tests

## Project Notes

- This is a **zmosh/zmx** session picker, not tmux-based. zmosh has its own command set (attach, kill, list, etc.) with no rename support.
- The picker suppresses zmosh command output and displays its own clean status messages.
- The TUI renders to `/dev/tty` and outputs only the final shell command to stdout, used via `eval "$(zpick)"`.
