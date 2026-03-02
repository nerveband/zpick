# zpick Release Procedure

Use this prompt when releasing a new version of zpick.

---

## Pre-Release Checklist

### 1. Code quality
```bash
make test          # all tests must pass
go vet ./...       # no vet warnings
make build         # binary compiles cleanly
```

### 2. Documentation updates
Before any release, verify these are current:

- **CHANGELOG.md** — new version section at top with bullet points for every user-facing change
- **README.md** — update if any of these changed:
  - CLI command table (the ``` block under "## CLI")
  - Install instructions or hook commands
  - Key bindings table
  - Session guard section
  - Backend list
- **`printUsage()` in `cmd/zp/main.go`** — must match README CLI table exactly
- **`assets/screenshot.svg`** — update if the TUI picker visuals changed (new keys in footer, new frame for a feature, layout changes). No update needed for CLI-only changes.

### 3. Version consistency
- CHANGELOG.md has the new version header
- Git tag will match (`vX.Y.Z`)
- No stale version references in docs

## Build & Local Install

```bash
make build         # builds ./zp with version from git tags
make install       # builds, copies to ~/.local/bin/zp, signs on macOS
```

`make install` handles macOS code signing automatically:
1. Copies binary to `~/.local/bin/zp`
2. Strips xattrs (`xattr -cr`) — needed because the repo lives in Dropbox
3. Ad-hoc signs (`codesign -fs -`) — without this, macOS SIGKILL's the binary
4. Creates `/usr/local/bin/zp` symlink if possible

### Verify local install
```bash
zp version         # should show current tag or dev
zp help            # verify command list is correct
```

## Release Flow

### 1. Stage, commit, tag
```bash
git add <files>
git commit -m "feat: description of changes"
git tag vX.Y.Z
git push origin main --tags
```

### 2. Run goreleaser
```bash
GITHUB_TOKEN=$(gh auth token) goreleaser release --clean
```

**Critical rules:**
- `GITHUB_TOKEN` is required — get it from `gh auth token`
- **Never** use bare `gh release create` — goreleaser attaches pre-built binaries + `checksums.txt` that `zp upgrade` (go-selfupdate with ChecksumValidator) depends on
- Goreleaser builds: `darwin/amd64`, `darwin/arm64`, `linux/amd64`, `linux/arm64`
- Goreleaser ships two binaries: `zp` (primary) and `zpick` (deprecation shim)

### 3. Upgrade local binary
```bash
make install       # rebuild from tagged source, sign, install
zp version         # confirm new version
```

### 4. Verify release
```bash
gh release view vX.Y.Z              # confirm binaries attached
gh release view vX.Y.Z --json assets | jq '.assets[].name'  # list all assets
```

Expected assets: `zp_X.Y.Z_{darwin,linux}_{amd64,arm64}.tar.gz` + `checksums.txt`

## When to Update the SVG

The `assets/screenshot.svg` is an animated SVG shown in the README. It cycles through 5 frames showing the TUI picker with different backends. Update it when:

- **New key added to the footer** (e.g., a new single-press key binding)
- **New backend frame** needed (currently: zmosh, tmux, zellij, shpool, in-session switching)
- **Layout changes** to the picker (column widths, status indicators, etc.)
- **Key mode** visuals changed (numbers vs letters labeling)

**Don't update** for CLI-only changes (new subcommands, flag changes, etc.) since the SVG shows the TUI, not the CLI.

## Cleanup After Release

Remove build artifacts that shouldn't linger:
```bash
rm -f zp zpick zmosh-picker
rm -rf dist
```

These are all gitignored but accumulate disk space (~45MB total).

## Troubleshooting

| Symptom | Fix |
|---------|-----|
| `zp` exits silently / SIGKILL | Re-sign: `xattr -cr ~/.local/bin/zp && codesign -fs - ~/.local/bin/zp` |
| `zp upgrade` fails | Check that release has `checksums.txt` — means goreleaser wasn't used |
| goreleaser fails "tag not found" | Ensure `git tag vX.Y.Z` was pushed: `git push origin vX.Y.Z` |
| goreleaser fails "token" | Set `GITHUB_TOKEN=$(gh auth token)` |
| Version shows "dev" or "dirty" | Commit all changes, then `git tag` on the clean commit |
