# Zellij Backend + Guard Cleanup

## Date: 2026-02-25

## Summary

Add zellij as a new backend to zpick and remove the zpick guard block from the user's `.zshrc`.

## Part 1: Remove zpick guard from `.zshrc`

Remove lines 283–303 — the `# >>> zpick guard >>>` block containing `zp()`, autorun, `_zpick_guard`, and the `codex`/`opencode` wrappers.

## Part 2: Zellij Backend

### New file: `internal/backend/zellij/zellij.go`

Follows the shpool pattern (simplest backend). Implements `backend.Backend`:

| Method | Implementation |
|---|---|
| `Name()` | `"zellij"` |
| `BinaryName()` | `"zellij"` |
| `SessionEnvVar()` | `"ZELLIJ"` |
| `InSession()` | `os.Getenv("ZELLIJ") != ""` |
| `Available()` | `exec.LookPath("zellij")` |
| `Version()` | `zellij --version` |
| `List()` | Parse `zellij list-sessions` |
| `FastList()` | Delegates to `List()` |
| `Attach()` | `syscall.Exec` zellij attach |
| `AttachCommand()` | `zellij attach <name>` |
| `Kill()` | `zellij kill-session <name>` |

### Zellij CLI details

- `zellij list-sessions` — one session per line, may include status markers like `(current session)` and `EXITED`
- `zellij attach <name>` — attach to existing session
- `zellij kill-session <name>` — kill a session
- `zellij --version` — version info
- Env vars: `ZELLIJ` (set to "0" inside session), `ZELLIJ_SESSION_NAME` (session name)

### Other file changes

- `internal/backend/types.go` — add `"ZELLIJ"` to `AllSessionEnvVars()`
- `internal/backend/config.go` — add `"zellij"` to `validBackends`
- `cmd/zp/main.go` — add blank import for zellij backend
- `internal/backend/zellij/zellij_test.go` — unit tests
