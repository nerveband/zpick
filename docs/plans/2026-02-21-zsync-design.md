# zsync — Native iOS SSH + zmosh Session Manager

## Problem

SSH from a phone is painful. You arrive on a remote machine and don't know what sessions exist, can't type long commands, and lose context every time the connection drops. zmosh-picker solves this on the Mac side — every terminal opens into a persistent session. But there's no native iOS experience to discover, manage, and attach to those sessions across machines.

zsync is "cmux for iOS" — a native app that SSHs into your machines, discovers zmosh sessions, and lets you attach with one tap. Machines sync via iCloud so your phone always knows where your sessions live.

## Design Decisions

### Pure Native Swift

SwiftUI + SwiftTerm + Citadel/swift-nio-ssh. No React Native, no Expo, no Blink fork. Terminal apps need precise text rendering, keyboard handling, and low-latency input — wrapper frameworks add lag. SwiftTerm (MIT, Miguel de Icaza) is the gold standard for Swift terminal emulation.

### SSH-First, Mosh Later

v1 ships SSH only. Mosh requires `libmoshios` (GPL-3.0 from Blink Shell), which would force the entire app to be GPL. Defer mosh support to v2 after resolving the licensing path. Consider Eternal Terminal (Apache-2.0) as an alternative.

### zmosh is Required

The app gates on zmosh. When connecting to a machine for the first time, zsync SSHs in, runs `zmosh list`, and checks if zmosh is installed. If not, it blocks access and shows a setup screen with installation instructions linking to the GitHub README. No bypass — zmosh is the entire point.

### iPhone-First

Design for iPhone. iPad and Mac can come later. The primary use case is picking up sessions from your phone while away from your desk.

## Architecture

### Tech Stack

| Component | Library | License |
|-----------|---------|---------|
| App shell | SwiftUI | Apple |
| Terminal emulation | SwiftTerm | MIT |
| SSH transport | Citadel + swift-nio-ssh | MIT / Apache-2.0 |
| Data persistence | SwiftData + CloudKit | Apple |
| Key storage | Keychain + Secure Enclave | Apple |
| Voice input | SFSpeechRecognizer (iOS 17+) | Apple |
| Mosh (v2) | libmoshios | GPL-3.0 (deferred) |

### Data Model

**Machine** (persisted, iCloud-synced via SwiftData + CloudKit):
- name, host, port, username
- authMethod (key or password), keyRef
- icon, iconColor
- zmoshInstalled (boolean, updated on each connect)

**Session** (transient, always live-queried via `zmosh list`):
- name, pid, clients, startedIn, isActive

**RecentSession** (persisted):
- sessionName, machineName, lastConnected

### Session Discovery

The app SSHs into each machine, runs `zmosh list`, and parses the output. Currently tab-separated text; should add `zmosh list --json` to the CLI for structured output before building the parser. Session data is never cached as stale — always re-fetched on screen load and on app wake.

## App Structure — Screens

### 1. Home

- **Quick Jump**: Vertical list of last 5 connected sessions across all machines. Each card shows session name, machine name, active/idle status, and time since last connect. Tap to attach immediately.
- **Machines**: List of configured machines with icon, name, session count, active count, and health dot. Machines without zmosh show a yellow "Setup required" badge.
- **Tab bar**: Sessions | Keys | Settings

### 2. Session Picker

Reached by tapping a machine (only if zmosh is verified). Shows all zmosh sessions on that machine:
- Each session is a tappable row with terminal icon, name, path, and active/idle indicator
- **New Session** button opens the new session flow
- Swipe left to reveal "Kill" action with confirmation sheet
- No numbered keys — this is iOS, not a TUI

### 3. Setup Required (gate screen)

Shown when connecting to a machine where zmosh is not installed. Blocks access entirely:
- Warning icon and "Setup Required" title
- 4 numbered installation steps
- "View Install Guide on GitHub" button (opens README)
- "Check Again" button (re-runs SSH + `zmosh list`)

### 4. New Session

Full-screen form consolidated from what were separate Pick Dir / +Date / Custom flows:
- **Name field**: Pre-filled with auto-generated name (directory-based with date or increment suffix). Tap X to clear and type custom name.
- **Directory picker**: Tap to open a searchable directory list (powered by zoxide on the remote). Picking a directory updates the session name.
- **Create** button in navbar

### 5. Terminal

SwiftTerm terminal view with:
- **Composition bar**: `[text input] [mic] [send]` — type or speak commands, then send. No Passthrough/Compose mode toggle (mode toggles cause confusion on mobile).
- **Key bar**: Scrollable row of special keys: snippets (star), history (clock), divider, esc, ctrl (sticky), tab, alt (sticky), divider, arrow keys, divider, ~, |, /, -, _
- **Snippet drawer**: Horizontally scrollable chips (git status, git pull, ls -la, npm run dev, docker ps, etc.)
- **History drawer**: Recent commands with timestamps

### 6. Add Machine

Form with:
- Display name, host, port, username
- Icon picker: 10 icons (laptop, desktop, server, cloud, chip, globe, terminal, network, home, secure) in a grid with distinct colors
- Auth method: SSH Key (default) or Password
- Yellow warning banner about zmosh being required

### 7. SSH Keys

- **Device Key (Secure Enclave)**: Ed25519 key generated on-device, never leaves the phone. Copy or share the public key to add to `~/.ssh/authorized_keys`.
- **Imported Keys**: Keys from iCloud Keychain or manual import.
- **Generate New Key** button.

### 8. Settings

- **Terminal Theme**: 8 themes (Dracula, Solarized Dark, Nord, Catppuccin Mocha, Tokyo Night, One Dark, Gruvbox Dark, Default) with live terminal preview showing selected font and theme.
- **Terminal Font**: 10 monospace fonts (JetBrains Mono default, SF Mono, Menlo, Fira Code, Source Code Pro, IBM Plex Mono, Cascadia Code, Hack, Inconsolata, Monaco). Each shows a sample in its own typeface.
- **Security**: Face ID lock toggle.
- **Sync**: iCloud Sync toggle, iCloud Keychain sync toggle.

## Voice Input

Two tiers:
1. **Keyboard dictation**: Free via UITextField mic button. Works out of the box but mangles CLI commands.
2. **SFSpeechRecognizer with Custom Language Model** (iOS 17+): Dedicated mic button in the composition bar. Trained on shell builtins, common flags, symbols ("pipe" -> `|`), git commands, file paths. On-device processing. Button shows recording state (red pulse animation).

Voice input always goes to the composition bar for review before sending. Never auto-sends to the terminal.

## Sync

SwiftData + CloudKit for machine configs. Zero backend maintenance. Syncs: machine name, host, port, username, auth method, icon, theme, font preferences.

**Does not sync**: Private keys (local by default), session data (always live-queried).

Private keys stay in the local Keychain unless the user explicitly marks them as synchronizable to iCloud Keychain.

## Technical Risks

1. **iOS background execution**: SSH connections die when suspended. Mosh (UDP) survives ~10min. Session Picker must refresh on every app wake — assume the SSH connection is dead.
2. **SwiftTerm + Citadel glue**: SwiftTerm expects a data stream; Citadel provides SSH channels. Bridging these requires a custom glue layer. Known issues with high-throughput SSH in SwiftTermApp.
3. **Host key trust UX**: First-connect TOFU (Trust On First Use) flow must be clear and secure. Show fingerprint, allow pinning, handle rotation.
4. **Session discovery latency**: SSHing into multiple machines serially is slow. Use parallel fetch + cache + timeout.
5. **CloudKit edge cases**: SwiftData + CloudKit has pitfalls around schema migration and iCloud-toggle behavior.

## Monetization

- **Free**: 1 machine, SSH only, no sync.
- **Pro** (one-time $19.99 or $19/year): Unlimited machines, iCloud sync, themes, fonts, voice input, snippets.
- Avoid subscription-only — terminal users value ownership.

## Scope

- **v1**: SSH + zmosh picker + keyboard + voice + themes/fonts + Secure Enclave keys + iCloud sync + dependency gate
- **v1.5**: External keyboard mode, snippet management, CloudKit polish
- **v2**: Mosh/ET transport (after licensing resolution), agent forwarding

## Prototype

Interactive HTML prototype: `prototype/zsync.html`

## Counselor Reviews

Independent reviews from Codex 5.3 XHigh and Gemini 3 Pro in `agents/counselors/1771684952-zsync-ios-design/`.
