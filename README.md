# zmosh-picker

Session launcher for [zmosh](https://github.com/mmonad/zmosh). One keypress to resume any session.

<p align="center">
  <img src="assets/screenshot.svg" alt="zmosh-picker in action" width="680">
</p>

## Why

The idea is simple: if every terminal on your Mac starts inside a zmosh session, then every session is always there waiting for you when you pick up your phone.

I SSH from my phone a lot. The annoying part was never the SSH itself — it was arriving on the remote machine and not knowing what sessions exist, or having to type `zmosh attach some-long-name` on a phone keyboard. Half the time I'd just start fresh and lose whatever I was working on.

So I made this run on every terminal. Now my Mac is constantly creating and resuming named sessions without me thinking about it. When I SSH from my phone later, those sessions are right there. Press `1`. Done.

It also handles new sessions — `Enter` creates one named after your current directory, `z` lets you pick a project via zoxide first. But the real point is: your Mac does the work of keeping sessions alive, and your phone just picks them up.

## Dependencies

| Dependency | Required | What it does |
|-----------|----------|---------|
| [zmosh](https://github.com/mmonad/zmosh) | **Yes** | Session persistence (fork of [zmx](https://github.com/neurosnap/zmx)) |
| [zoxide](https://github.com/ajeetdsouza/zoxide) | No | Directory picker for the `z` key |
| [fzf](https://github.com/junegunn/fzf) | No | Fuzzy finder, used by zoxide |

```bash
# macOS
brew install mmonad/tap/zmosh
brew install zoxide    # optional
brew install fzf       # optional, used by zoxide
```

## Install

### From source (Go 1.22+)

```bash
go install github.com/nerveband/zmosh-picker/cmd/zmosh-picker@latest
```

### From GitHub releases

```bash
curl -fsSL https://raw.githubusercontent.com/nerveband/zmosh-picker/main/install.sh | bash
```

### Build from source

```bash
git clone https://github.com/nerveband/zmosh-picker.git
cd zmosh-picker
make install
```

### Add the shell hook

```bash
zmosh-picker install-hook
```

This adds the hook line to your `.zshrc` (or `.bashrc`). If you use Powerlevel10k, it places the hook before instant prompt so the picker can read keyboard input.

Open a new terminal and you should see it.

## CLI

```
zmosh-picker              Interactive TUI picker (default)
zmosh-picker list         List sessions (human-readable)
zmosh-picker list --json  List sessions (JSON for scripts/zsync)
zmosh-picker check        Check dependencies
zmosh-picker check --json Machine-readable dependency check
zmosh-picker attach <n>   Attach or create session
zmosh-picker kill <name>  Kill a session
zmosh-picker install-hook Add shell hook to .zshrc/.bashrc
zmosh-picker upgrade      Upgrade to the latest version
zmosh-picker version      Print version
```

### `list --json` output

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

### `check --json` output

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

## Usage

```
  zmosh 3 sessions

  1  api-server * ~/projects/api-server
  2  dotfiles . ~/dotfiles
  3  ai-happy-design . ~/Doc/GH/ai-happy-design

  enter new myproject
  c custom  z pick dir  d +date  k kill  esc skip

  >
```

### Keys

| Key | What happens |
|-----|--------|
| `1`-`9` | Attach to that session |
| `a`-`y` | Sessions 10+ |
| `Enter` | New session in current directory |
| `c` | Custom name — type a name, then pick where to create it |
| `z` | Pick a directory with zoxide, then new session there |
| `d` | New session with today's date as suffix |
| `k` | Kill mode — pick a session to kill (returns to menu after) |
| `Esc` | Skip, just give me a normal shell |

Everything is single-press. No typing names, no confirming.

### Session names

| Key | Format | Example |
|-----|--------|---------|
| `Enter` | `<dirname>` or `<dirname>-<N>` | `api-server`, then `api-server-2` |
| `c` | whatever you type | `my-thing` |
| `d` | `<dirname>-MMDD` | `api-server-0220` |
| `z` | `<picked-dir>` or `<picked-dir>-<N>` | `ai-happy-design`, then `ai-happy-design-2` |

First session gets the bare name. Counter starts at `-2` only when a conflict exists.

### The `*` and `.` indicators

`*` (green) means someone is connected to that session right now. Probably you, on another device. `.` means it's idle — pick it up.

## How it works

The `.zshrc` hook runs the binary before the rest of your shell config. It calls `zmosh list` once, builds the menu, waits for one keypress, then runs `exec zmosh attach <name>` which replaces the process entirely. Resuming a session is actually faster than a normal shell startup.

The picker skips itself when you're already inside a zmosh session, in a non-interactive shell, or when stdin isn't a terminal.

### Killing sessions

Press `k` to enter kill mode, then pick a session number to kill. You'll be asked to confirm with `y/n`. To skip confirmation:

```bash
export ZMOSH_PICKER_NO_CONFIRM=1
```

## Cross-platform

zmosh-picker is a Go binary that works on:
- macOS (arm64, amd64)
- Linux (arm64, amd64)
- zsh and bash

## Works on narrow screens

The layout fits ~30 character widths. Action keys are stacked on two lines. No padding on session names. I built this mostly so I could SSH from my phone and not hate the experience.

## Uninstall

```bash
zmosh-picker install-hook --remove
```

## Related projects

- [zmosh](https://github.com/mmonad/zmosh) — Session persistence with UDP remote support
- [zmx](https://github.com/neurosnap/zmx) — The session persistence tool zmosh is forked from
- [zmx-session-manager](https://github.com/mdsakalu/zmx-session-manager) — TUI session manager for zmx/zmosh
- [zoxide](https://github.com/ajeetdsouza/zoxide) — Frecency-based `cd` replacement
- [fzf](https://github.com/junegunn/fzf) — Fuzzy finder

## License

MIT
