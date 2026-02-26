# In-Session Switching, PATH Symlink, and SVG Update

**Date:** 2026-02-26

## Problem

1. Running `zp` inside an existing session just prints a warning and exits. Users must remember backend-specific detach commands and keyboard shortcuts to switch sessions.
2. `zp` installs to `~/.local/bin/` which isn't in the default system PATH. This breaks `mosh host -- zp` and other non-interactive contexts.
3. The README SVG animation doesn't showcase the in-session switching feature.

## Feature 1: In-Session Picker

### Behavior

When `zp` detects `InSession()`, show the full picker instead of the warning message. The current session is marked with `←` in the session list.

**Picker display (in-session mode):**

```
zellij 3 sessions  (in: frontend ←)

1  frontend  ← ~/projects/frontend
2  backend   .  ~/projects/backend
3  scratch   .  ~/tmp

enter new   myproject
c custom  z pick dir  d +date
k kill  h help  esc skip

>
```

### Actions

All existing picker actions work. The difference is in how attach/create actions are fulfilled:

- **Attach/New/Custom/Zoxide/Date** — Write target to file, output detach command
- **Kill** — Works normally (no detach needed, just kills the target session)
- **Escape** — Cancel, stay in current session

### Detach + Auto-Relaunch Mechanism

1. User selects session "backend" from inside session "frontend"
2. `zp` writes action to `~/.cache/zpick/switch-target` as JSON:
   ```json
   {"action":"attach","name":"backend"}
   ```
   Or for new sessions:
   ```json
   {"action":"new","name":"scratch-0226","dir":"/tmp"}
   ```
3. `zp` outputs the backend's detach command to stdout (via `eval`)
4. Terminal detaches from session, terminal emulator spawns new shell
5. Shell hook (precmd) detects `~/.cache/zpick/switch-target`, runs `zp resume`
6. `zp resume` reads file, deletes it, outputs the attach command
7. User lands in "backend" session

### Backend Interface Addition

```go
// DetachCommand returns the shell command to detach from the current session.
DetachCommand() string
```

| Backend | DetachCommand |
|---------|--------------|
| tmux    | `tmux detach-client` |
| zellij  | `zellij action detach` |
| zmx     | `zmx detach` |
| zmosh   | `zmx detach` |
| shpool  | `shpool detach` |

### New Subcommand: `zp resume`

Reads `~/.cache/zpick/switch-target`, deletes the file, and outputs the appropriate shell command (attach or create+attach) to stdout. If the file doesn't exist or is stale (>30s old), outputs nothing.

### Shell Hook Addition

Added to the guard block, using the same precmd pattern as the existing autorun hook:

```bash
if [[ -f ~/.cache/zpick/switch-target ]]; then
  _zpick_switch() {
    precmd_functions=(${precmd_functions:#_zpick_switch})
    eval "$(command zp resume)"
  }
  precmd_functions+=(_zpick_switch)
fi
```

Fish equivalent added to `zp.fish`.

## Feature 2: PATH Symlink (`/usr/local/bin/zp`)

### Problem

`~/.local/bin` is not in the default system PATH. This breaks:
- `mosh host -- zp`
- Cron jobs
- Any non-interactive context that doesn't source shell profiles

### Solution

Create a symlink: `/usr/local/bin/zp` -> `~/.local/bin/zp`

### Touch Points

1. **`make install`** — Creates symlink after building/signing. If `/usr/local/bin` isn't writable, prints sudo hint.
2. **`zp install-hook`** — Also creates the symlink (this is the setup command for prebuilt binary users).
3. **`zp upgrade`** — Does NOT create or recreate the symlink. If symlink doesn't exist, prints: `note: /usr/local/bin/zp not found — run 'zp install-hook' to add it`

### User Intent

If a user manually removes the symlink, it stays gone. Only explicit user actions (`install-hook`, `make install`) create it. `upgrade` only informs, never forces.

## Feature 3: SVG Animation Update

### Current State

4-frame animation cycling: zmosh, tmux, zellij, shpool. Each frame shows the standard picker view.

### Change

Add a 5th frame showing the in-session picker with the `←` marker on the current session, demonstrating the switching capability. Adjust keyframe timing from 4-frame to 5-frame cycle. Update tagline to mention session switching.
