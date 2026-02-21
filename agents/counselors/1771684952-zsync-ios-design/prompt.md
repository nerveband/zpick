# Design Review Request: Zsync — Native iOS SSH/Mosh Session Manager

## Question

We are designing a native iOS app called **Zsync** — a zmosh session picker and SSH terminal client. Think "cmux for iOS" — a native app that discovers, manages, and connects to zmosh sessions across multiple machines, synced via iCloud.

Please review this design for:
- Architectural soundness
- Risks and blind spots
- UX considerations for a mobile terminal app
- Technical feasibility
- Missing features or overlooked concerns
- Whether the tech stack choices are optimal

## Context

### What is zmosh-picker (the existing project)?

Read @zmosh-picker to understand the current zsh-based session picker. It's a pure-zsh, single-keypress TUI that shows zmosh sessions on login and lets you attach with one keypress. Mobile-first design — optimized for SSH from phones.

Read @README.md for the full project description and design philosophy.

Read @docs/plans/2026-02-20-zmosh-picker-design.md for the design rationale.

### What is zmosh?

zmosh is a fork of zmx — a persistent terminal session manager for macOS. It provides:
- Named session creation/attachment
- Client tracking (who's connected)
- UDP remote support (sessions accessible over mosh)
- `zmosh list` outputs tab-separated fields: session_name, pid, clients, created_at, started_in

### The Zsync iOS App Design

**Architecture: Pure Native Swift**
- SwiftUI app shell
- SwiftTerm (MIT, by Miguel de Icaza) for terminal emulation — provides a native UIKit iOSTerminalView
- Citadel (MIT) + swift-nio-ssh (Apache-2.0, Apple) for pure-Swift SSH
- SwiftData + CloudKit for syncing machine configs across devices
- Keychain + iCloud Keychain for SSH keys
- SFSpeechRecognizer with custom terminal vocabulary for voice input
- libmoshios from Blink Shell for mosh support (future — GPL-3.0 licensing consideration)

**App Structure — Three Screens:**

1. **Home (Machines list)**
   - "Recent Sessions" — horizontally scrollable cards showing last 5 connected sessions regardless of machine (quick jump)
   - Machine list — tap to see sessions on that machine
   - iCloud-synced machine configurations

2. **Session Picker**
   - Shows zmosh sessions on selected machine (parsed from `zmosh list` via SSH)
   - Each session shows: name, path, active/idle status
   - Action buttons: New, Custom name, Pick Dir, +Date (matching zmosh-picker's keybindings)
   - Swipe-to-kill gesture

3. **Terminal**
   - SwiftTerm terminal view
   - Mode toggle: Passthrough (keystrokes go directly to terminal) vs Compose (type full command then send)
   - Composition bar: UITextField with iOS dictation support + dedicated mic button using SFSpeechRecognizer with custom CLI vocabulary
   - Key bar: esc, ctrl (sticky toggle), tab, alt (sticky toggle), arrow keys, ~, |, /, -, _
   - Built-in color schemes (Dracula, Solarized Dark, Nord, Catppuccin, Tokyo Night, One Dark, Gruvbox, etc.)

**Data Model:**
- Machine (persisted, synced): name, host, port, username, authMethod, keyRef
- Session (transient, always live-queried via `zmosh list`): name, clients, startedIn, isActive
- RecentSession (persisted): sessionName, machineName, lastConnected (for quick jump)

**Voice Input:**
- Composition bar is a standard UITextField — gets iOS keyboard dictation mic for free
- Dedicated mic button uses SFSpeechRecognizer with custom language model (iOS 17+) trained on terminal commands
- Custom vocab includes: common CLI commands, file path patterns, git commands, shell syntax

**Key Technical Challenges:**
- iOS background execution: mosh/UDP connections survive ~10min in background, then reconnect instantly (acceptable for zmosh)
- Terminal keyboard input: SwiftTerm's iOSTerminalView handles UIKeyInput; composition bar is a separate UITextField
- Session discovery: app SSHs into each machine, runs `zmosh list`, parses output

**What we chose NOT to do:**
- Fork Blink Shell (GPL-3.0 would infect the whole app)
- Use React Native/Expo (double bridge overhead, weak terminal ecosystem)
- Use libghostty (premature — no iOS target, C API not released yet)
- Build for iPad/Mac initially (iPhone-first, add later)

### Interactive Prototype
Read @prototype/zsync.html to see the interactive HTML prototype of the UI.

## Instructions

You are providing an independent review of this iOS app design. Be critical and thorough.

1. **Architecture**: Is the pure-native Swift stack the right call? Are there risks with SwiftTerm, Citadel, or swift-nio-ssh we should know about?
2. **UX**: Does the three-screen flow make sense? Is the quick-jump feature well-designed? Is the composition bar / key bar the right approach for mobile terminal input?
3. **Technical Risks**: What are the biggest technical risks? Background execution, SSH key management, session discovery latency, SwiftTerm integration challenges?
4. **Missing Features**: What's missing that a v1 iOS terminal app needs? What would users expect that we haven't considered?
5. **Licensing**: The Blink Shell mosh library (libmoshios) is GPL-3.0. What are the implications? Are there alternatives for mosh on iOS?
6. **Voice Input**: Is SFSpeechRecognizer with custom vocabulary the right approach? Is it worth the extra code over just the keyboard mic?
7. **Sync**: Is SwiftData + CloudKit the right choice for syncing machine configs? Any pitfalls?
8. **Monetization**: How should this app be monetized? Free with IAP? Paid upfront? Subscription?

Be direct and opinionated. Suggest alternatives where you see better approaches. Structure your response with clear headings.
