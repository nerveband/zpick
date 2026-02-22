# zpick

Session launcher for [zmosh](https://github.com/mmonad/zmosh). One keypress to resume any session.

<p align="center">
  <img src="assets/screenshot.svg" alt="zpick in action" width="680">
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
go install github.com/nerveband/zpick/cmd/zpick@latest
```

### From GitHub releases

```bash
curl -fsSL https://raw.githubusercontent.com/nerveband/zpick/main/install.sh | bash
```

### Build from source

```bash
git clone https://github.com/nerveband/zpick.git
cd zpick
make install
```

### Add the shell hook

```bash
zpick install-hook
```

This auto-detects your shell (zsh or bash) and adds the hook line to your `.zshrc` or `.bashrc`. If you use Powerlevel10k, it places the hook before instant prompt so the picker can read keyboard input.

Open a new terminal and you should see it.

#### Manual hook setup

If you prefer to add the hook manually, add this to your shell config:

```bash
# zpick: session launcher
[[ -z "$ZMX_SESSION" ]] && command -v zpick &>/dev/null && eval "$(zpick)"
```

## CLI

```
zpick              Interactive TUI picker (default)
zpick list         List sessions (human-readable)
zpick list --json  List sessions (JSON for scripts)
zpick check        Check dependencies
zpick check --json Machine-readable dependency check
zpick attach <n>   Attach or create session
zpick kill <name>  Kill a session
zpick install-hook Add shell hook to .zshrc/.bashrc
zpick upgrade      Upgrade to the latest version
zpick version      Print version
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

## How it works

The TUI renders to `/dev/tty` so it works even when stdout is piped. Only the final shell command goes to stdout, which the hook `eval`s. This means:

- `eval "$(zpick)"` — TUI appears, selecting a session runs `exec zmosh attach <name>`
- Press Escape — empty output, shell prompt returns normally
- `zpick | cat` — only the command string appears, TUI still renders on the terminal

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

### Killing sessions

Press `k` to enter kill mode, then pick a session number to kill. You'll be asked to confirm with `y/n`. To skip confirmation:

```bash
export ZPICK_NO_CONFIRM=1
```

## Cross-platform

zpick is a Go binary that works on:
- macOS (arm64, amd64)
- Linux (arm64, amd64)
- zsh and bash

## Works on narrow screens

The layout fits ~30 character widths. Action keys are stacked on two lines. No padding on session names. I built this mostly so I could SSH from my phone and not hate the experience.

## Uninstall

```bash
zpick install-hook --remove
```

## Related projects

- [zmosh](https://github.com/mmonad/zmosh) — Session persistence with UDP remote support
- [zmx](https://github.com/neurosnap/zmx) — The session persistence tool zmosh is forked from
- [zmx-session-manager](https://github.com/mdsakalu/zmx-session-manager) — TUI session manager for zmx/zmosh
- [zoxide](https://github.com/ajeetdsouza/zoxide) — Frecency-based `cd` replacement
- [fzf](https://github.com/junegunn/fzf) — Fuzzy finder

## License

MIT
