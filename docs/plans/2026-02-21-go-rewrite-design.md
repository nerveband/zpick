# zmosh-picker v2: Go Rewrite

## Problem

The current zmosh-picker is a pure zsh script. That's fine for macOS + zsh, but it:
- Only works on zsh (no bash, fish, Linux server support)
- Has no structured output (zsync iOS app needs `--json`)
- Has no tests
- Can't handle OS-specific quirks cleanly
- Requires `source` (can't be a standalone binary)
- Has no dependency checking or helpful error messages

## Solution

Rewrite zmosh-picker as a Go CLI. Single static binary, cross-platform, testable, with structured output for zsync.

## Design Decisions

### Go over Rust/Python/Node

- Go compiles to a single static binary — no runtime, no dependencies
- Cross-compiles trivially: `GOOS=linux GOARCH=arm64 go build`
- Fast startup (~20ms), critical for a tool that runs on every terminal open
- Excellent terminal libraries (bubbletea, x/term)
- `go install` for users, `brew install` for macOS

### Same UX, Better Foundation

The single-keypress picker UX stays identical. Users shouldn't notice the rewrite except that it now works on Linux and has better error messages.

### Subcommands for Programmatic Use

Default (no subcommand) = interactive picker (same as today). Subcommands add machine-readable interfaces for zsync and scripts.

## CLI Interface

```
zmosh-picker              # Interactive TUI picker (default)
zmosh-picker list         # List sessions (human-readable table)
zmosh-picker list --json  # List sessions (JSON for zsync)
zmosh-picker check        # Check if zmosh is installed, report status
zmosh-picker check --json # Machine-readable dependency check
zmosh-picker attach <name> [--dir <path>]  # Attach/create session
zmosh-picker kill <name>  # Kill a session
zmosh-picker version      # Print version
zmosh-picker install-hook # Add .zshrc/.bashrc hook
```

### `list --json` Output

```json
{
  "sessions": [
    {
      "name": "bbcli",
      "pid": 5678,
      "clients": 1,
      "started_in": "~/Documents/GitHub/agent-to-bricks",
      "active": true
    }
  ],
  "count": 1,
  "zmosh_version": "0.4.2"
}
```

### `check --json` Output

```json
{
  "zmosh": {"installed": true, "version": "0.4.2", "path": "/opt/homebrew/bin/zmosh"},
  "zoxide": {"installed": true, "version": "0.9.4", "path": "/opt/homebrew/bin/zoxide"},
  "fzf": {"installed": true, "version": "0.46.0", "path": "/opt/homebrew/bin/fzf"},
  "shell": "zsh",
  "os": "darwin",
  "arch": "arm64"
}
```

## Architecture

```
zmosh-picker/
├── cmd/
│   └── zmosh-picker/
│       └── main.go              # CLI entry point
├── internal/
│   ├── zmosh/
│   │   ├── parser.go            # Parse `zmosh list` output
│   │   ├── parser_test.go
│   │   ├── session.go           # Session operations (attach, kill, list)
│   │   └── session_test.go
│   ├── picker/
│   │   ├── picker.go            # Interactive TUI picker
│   │   ├── picker_test.go
│   │   ├── keys.go              # Key mapping (1-9, a-y)
│   │   ├── keys_test.go
│   │   ├── naming.go            # Name generation (counter, date)
│   │   └── naming_test.go
│   ├── check/
│   │   ├── deps.go              # Dependency checking
│   │   └── deps_test.go
│   └── hook/
│       ├── install.go           # Shell hook installation
│       ├── install_test.go
│       ├── zsh.go               # zsh-specific hook
│       └── bash.go              # bash-specific hook
├── go.mod
├── go.sum
├── Makefile
├── install.sh                   # Updated: downloads binary or runs go install
├── README.md                    # Updated for Go version
├── docs/plans/
│   └── ...
└── .goreleaser.yml              # Cross-platform release builds
```

## Dependencies

Minimal:
- `golang.org/x/term` — raw terminal mode for single-keypress input
- No TUI framework (bubbletea is overkill for this simple picker)
- Standard library for everything else (os/exec, encoding/json, fmt)

## Key Behaviors

### Interactive Picker (default mode)

Exact same UX as current zsh version:
1. Run `zmosh list`, parse output
2. Display numbered session list with status indicators
3. Show action keys at bottom (Enter, z, d, c, k, Esc)
4. Wait for single keypress
5. Execute action (attach, create, kill, etc.)
6. Loop back to list after kill

### Session Naming

Ported from zsh:
- **Counter mode**: `<dirname>` → `<dirname>-2` → `<dirname>-3`
- **Date mode**: `<dirname>-MMDD`
- Base name from `basename(cwd)` or zoxide-picked directory

### Shell Hook

The `.zshrc` / `.bashrc` hook:
```sh
# zmosh-picker: session launcher
[[ -z "$ZMX_SESSION" ]] && command -v zmosh-picker &>/dev/null && zmosh-picker
```

Same as today, but now it calls a compiled binary instead of sourcing a script.

`zmosh-picker install-hook` automates adding this line (detects shell, handles p10k placement for zsh).

### Cross-Platform Handling

Build tags for OS-specific code:
- `hook/zsh.go` — zsh hook format, p10k detection
- `hook/bash.go` — bash hook format, .bashrc vs .bash_profile
- Future: `hook/fish.go`

Runtime detection:
- `$SHELL` for current shell
- `runtime.GOOS` / `runtime.GOARCH` for platform

### Error Messages

Replace `2>/dev/null` with helpful messages:
```
zmosh-picker: zmosh not found
  Install: brew install mmonad/tap/zmosh
  More info: https://github.com/nerveband/zmosh-picker#dependencies
```

## Distribution

1. **go install**: `go install github.com/nerveband/zmosh-picker/cmd/zmosh-picker@latest`
2. **brew**: `brew install nerveband/tap/zmosh-picker`
3. **GitHub Releases**: Pre-built binaries via goreleaser (darwin/arm64, darwin/amd64, linux/arm64, linux/amd64)
4. **install.sh**: Updated script that downloads the right binary or falls back to `go install`

## Migration from v1

The Go binary is a drop-in replacement. The `.zshrc` hook line is identical. Users can:
1. `brew upgrade zmosh-picker` (once the tap is updated)
2. Or: `go install github.com/nerveband/zmosh-picker/cmd/zmosh-picker@latest`

The old zsh script in `~/.local/bin/zmosh-picker` gets replaced by the Go binary.
