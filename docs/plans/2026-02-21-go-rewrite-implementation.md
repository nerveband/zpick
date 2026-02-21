# zmosh-picker Go Rewrite Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Rewrite zmosh-picker from a zsh script to a Go CLI binary — same UX, cross-platform, testable, with `--json` output for the zsync iOS app.

**Architecture:** Single Go binary with subcommands. `internal/` packages for zmosh parsing, interactive picker, dependency checking, and shell hook management. No TUI framework — raw terminal via `golang.org/x/term` to keep it minimal and fast.

**Tech Stack:** Go 1.22+, golang.org/x/term, standard library (os/exec, encoding/json)

**Design Doc:** `docs/plans/2026-02-21-go-rewrite-design.md`

**Reference:** The current zsh script at `zmosh-picker` (repo root) defines the exact UX to replicate.

---

## Phase 1: Project Setup

### Task 1: Initialize Go Module

**Files:**
- Create: `go.mod`
- Create: `cmd/zmosh-picker/main.go`
- Keep: `zmosh-picker` (old zsh script, renamed to `zmosh-picker.zsh` for reference)

**Step 1: Rename old script for reference**

```bash
mv zmosh-picker zmosh-picker.zsh
```

**Step 2: Initialize Go module**

```bash
go mod init github.com/nerveband/zmosh-picker
```

**Step 3: Create main.go with subcommand routing**

```go
// cmd/zmosh-picker/main.go
package main

import (
	"fmt"
	"os"
)

var version = "dev"

func main() {
	if len(os.Args) < 2 {
		// Default: interactive picker
		if err := runPicker(); err != nil {
			fmt.Fprintf(os.Stderr, "zmosh-picker: %v\n", err)
			os.Exit(1)
		}
		return
	}

	switch os.Args[1] {
	case "list":
		if err := runList(); err != nil {
			fmt.Fprintf(os.Stderr, "zmosh-picker: %v\n", err)
			os.Exit(1)
		}
	case "check":
		if err := runCheck(); err != nil {
			fmt.Fprintf(os.Stderr, "zmosh-picker: %v\n", err)
			os.Exit(1)
		}
	case "attach":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: zmosh-picker attach <name> [--dir <path>]")
			os.Exit(1)
		}
		if err := runAttach(os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "zmosh-picker: %v\n", err)
			os.Exit(1)
		}
	case "kill":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: zmosh-picker kill <name>")
			os.Exit(1)
		}
		if err := runKill(os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "zmosh-picker: %v\n", err)
			os.Exit(1)
		}
	case "install-hook":
		if err := runInstallHook(); err != nil {
			fmt.Fprintf(os.Stderr, "zmosh-picker: %v\n", err)
			os.Exit(1)
		}
	case "version":
		fmt.Printf("zmosh-picker %s\n", version)
	case "--help", "-h", "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "zmosh-picker: unknown command %q\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

// Stubs — implemented in later tasks
func runPicker() error    { fmt.Println("picker not yet implemented"); return nil }
func runList() error      { fmt.Println("list not yet implemented"); return nil }
func runCheck() error     { fmt.Println("check not yet implemented"); return nil }
func runAttach(name string) error { fmt.Printf("attach %s not yet implemented\n", name); return nil }
func runKill(name string) error   { fmt.Printf("kill %s not yet implemented\n", name); return nil }
func runInstallHook() error { fmt.Println("install-hook not yet implemented"); return nil }

func printUsage() {
	fmt.Println(`zmosh-picker — session launcher for zmosh

Usage:
  zmosh-picker              Interactive TUI picker (default)
  zmosh-picker list         List sessions (--json for machine-readable)
  zmosh-picker check        Check dependencies (--json for machine-readable)
  zmosh-picker attach <n>   Attach or create session
  zmosh-picker kill <name>  Kill a session
  zmosh-picker install-hook Add shell hook to .zshrc/.bashrc
  zmosh-picker version      Print version`)
}
```

**Step 4: Build and test**

```bash
go build -o zmosh-picker ./cmd/zmosh-picker
./zmosh-picker version
./zmosh-picker --help
```

Expected: Prints version and usage

**Step 5: Commit**

```bash
git add go.mod cmd/ zmosh-picker.zsh
git rm zmosh-picker
git commit -m "feat: initialize Go module with CLI subcommand routing"
```

---

## Phase 2: Core — zmosh Parser

### Task 2: Parse `zmosh list` Output

**Files:**
- Create: `internal/zmosh/parser.go`
- Create: `internal/zmosh/parser_test.go`
- Create: `internal/zmosh/types.go`

**Step 1: Define types**

```go
// internal/zmosh/types.go
package zmosh

// Session represents a zmosh session from `zmosh list` output.
type Session struct {
	Name      string `json:"name"`
	PID       int    `json:"pid,omitempty"`
	Clients   int    `json:"clients"`
	StartedIn string `json:"started_in"`
	Active    bool   `json:"active"`
}

// ListResult is the JSON output format for `zmosh-picker list --json`.
type ListResult struct {
	Sessions     []Session `json:"sessions"`
	Count        int       `json:"count"`
	ZmoshVersion string    `json:"zmosh_version,omitempty"`
}
```

**Step 2: Write failing tests**

```go
// internal/zmosh/parser_test.go
package zmosh

import (
	"testing"
)

func TestParseSessions(t *testing.T) {
	input := "session_name=apcsp-1\tpid=1234\tclients=1\tstarted_in=~/GitHub/aak-class-25-26/apcsp\n" +
		"session_name=bbcli\tpid=5678\tclients=0\tstarted_in=~/Documents/GitHub/agent-to-bricks\n"

	sessions := ParseSessions(input)

	if len(sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(sessions))
	}
	if sessions[0].Name != "apcsp-1" {
		t.Errorf("expected name apcsp-1, got %s", sessions[0].Name)
	}
	if sessions[0].Clients != 1 {
		t.Errorf("expected 1 client, got %d", sessions[0].Clients)
	}
	if !sessions[0].Active {
		t.Error("expected session to be active")
	}
	if sessions[1].Name != "bbcli" {
		t.Errorf("expected name bbcli, got %s", sessions[1].Name)
	}
	if sessions[1].Active {
		t.Error("expected session to be idle")
	}
}

func TestParseEmpty(t *testing.T) {
	sessions := ParseSessions("")
	if len(sessions) != 0 {
		t.Fatalf("expected 0 sessions, got %d", len(sessions))
	}
}

func TestParseMissingFields(t *testing.T) {
	input := "session_name=test\tclients=0\n"
	sessions := ParseSessions(input)
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}
	if sessions[0].StartedIn != "~" {
		t.Errorf("expected default startedIn ~, got %s", sessions[0].StartedIn)
	}
}

func TestParseSkipsBlankLines(t *testing.T) {
	input := "\n\nsession_name=test\tclients=1\tstarted_in=~/foo\n\n"
	sessions := ParseSessions(input)
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}
}
```

**Step 3: Run tests — verify they fail**

```bash
go test ./internal/zmosh/ -v
```
Expected: Compilation error — ParseSessions doesn't exist

**Step 4: Implement parser**

```go
// internal/zmosh/parser.go
package zmosh

import (
	"strconv"
	"strings"
)

// ParseSessions parses the tab-separated output of `zmosh list`.
// Each line has fields like: session_name=foo\tpid=123\tclients=1\tstarted_in=~/bar
func ParseSessions(output string) []Session {
	var sessions []Session

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var s Session
		s.StartedIn = "~" // default

		for _, field := range strings.Split(line, "\t") {
			field = strings.TrimSpace(field)
			if k, v, ok := strings.Cut(field, "="); ok {
				switch k {
				case "session_name":
					s.Name = v
				case "pid":
					s.PID, _ = strconv.Atoi(v)
				case "clients":
					s.Clients, _ = strconv.Atoi(v)
				case "started_in":
					s.StartedIn = v
				}
			}
		}

		if s.Name == "" {
			continue
		}
		s.Active = s.Clients > 0
		sessions = append(sessions, s)
	}

	return sessions
}
```

**Step 5: Run tests — verify they pass**

```bash
go test ./internal/zmosh/ -v
```
Expected: All 4 tests PASS

**Step 6: Commit**

```bash
git add internal/zmosh/
git commit -m "feat: zmosh list parser with full test coverage"
```

---

### Task 3: Session Operations (list, attach, kill)

**Files:**
- Create: `internal/zmosh/session.go`
- Create: `internal/zmosh/session_test.go`

**Step 1: Write tests**

```go
// internal/zmosh/session_test.go
package zmosh

import (
	"testing"
)

func TestAttachCommand(t *testing.T) {
	cmd := AttachCommand("my-session")
	if cmd != `zmosh attach "my-session"` {
		t.Errorf("unexpected command: %s", cmd)
	}
}

func TestKillCommand(t *testing.T) {
	cmd := KillCommand("my-session")
	if cmd != `zmosh kill "my-session"` {
		t.Errorf("unexpected command: %s", cmd)
	}
}

func TestListCommand(t *testing.T) {
	cmd := ListCommand()
	if cmd != "zmosh list" {
		t.Errorf("unexpected command: %s", cmd)
	}
}
```

**Step 2: Implement session operations**

```go
// internal/zmosh/session.go
package zmosh

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// ListCommand returns the command string to list sessions.
func ListCommand() string {
	return "zmosh list"
}

// AttachCommand returns the command string to attach to a session.
func AttachCommand(name string) string {
	return fmt.Sprintf(`zmosh attach "%s"`, name)
}

// KillCommand returns the command string to kill a session.
func KillCommand(name string) string {
	return fmt.Sprintf(`zmosh kill "%s"`, name)
}

// List runs `zmosh list` and returns parsed sessions.
func List() ([]Session, error) {
	out, err := exec.Command("zmosh", "list").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run zmosh list: %w", err)
	}
	return ParseSessions(string(out)), nil
}

// ListJSON runs `zmosh list` and returns JSON.
func ListJSON() (string, error) {
	sessions, err := List()
	if err != nil {
		return "", err
	}
	result := ListResult{
		Sessions: sessions,
		Count:    len(sessions),
	}
	// Try to get zmosh version
	if v, err := exec.Command("zmosh", "version").Output(); err == nil {
		result.ZmoshVersion = strings.TrimSpace(string(v))
	}
	b, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Attach replaces the current process with `zmosh attach <name>`.
// Uses exec (syscall.Exec) so the picker process is replaced entirely,
// matching the zsh script's `exec zmosh attach` behavior.
func Attach(name string) error {
	zmoshPath, err := exec.LookPath("zmosh")
	if err != nil {
		return fmt.Errorf("zmosh not found: %w", err)
	}
	return syscall.Exec(zmoshPath, []string{"zmosh", "attach", name}, os.Environ())
}

// AttachInDir changes to dir first, then attaches.
func AttachInDir(name, dir string) error {
	if dir != "" {
		if err := os.Chdir(dir); err != nil {
			return fmt.Errorf("failed to cd to %s: %w", dir, err)
		}
	}
	return Attach(name)
}

// Kill runs `zmosh kill <name>`.
func Kill(name string) error {
	cmd := exec.Command("zmosh", "kill", name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
```

**Step 3: Run tests, commit**

```bash
go test ./internal/zmosh/ -v
git add internal/zmosh/session.go internal/zmosh/session_test.go
git commit -m "feat: session operations — list, attach (with exec), kill"
```

---

### Task 4: Session Naming

**Files:**
- Create: `internal/picker/naming.go`
- Create: `internal/picker/naming_test.go`

**Step 1: Write tests**

```go
// internal/picker/naming_test.go
package picker

import (
	"testing"

	"github.com/nerveband/zmosh-picker/internal/zmosh"
)

func TestCounterName_NoConflict(t *testing.T) {
	name := CounterName("projects", nil)
	if name != "projects" {
		t.Errorf("expected 'projects', got '%s'", name)
	}
}

func TestCounterName_WithConflict(t *testing.T) {
	existing := []zmosh.Session{
		{Name: "projects"},
	}
	name := CounterName("projects", existing)
	if name != "projects-2" {
		t.Errorf("expected 'projects-2', got '%s'", name)
	}
}

func TestCounterName_MultipleConflicts(t *testing.T) {
	existing := []zmosh.Session{
		{Name: "projects"},
		{Name: "projects-2"},
		{Name: "projects-3"},
	}
	name := CounterName("projects", existing)
	if name != "projects-4" {
		t.Errorf("expected 'projects-4', got '%s'", name)
	}
}

func TestDateName(t *testing.T) {
	name := DateName("projects")
	// Should match projects-MMDD pattern
	if len(name) < 10 {
		t.Errorf("expected projects-MMDD format, got '%s'", name)
	}
}
```

**Step 2: Implement naming**

```go
// internal/picker/naming.go
package picker

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/nerveband/zmosh-picker/internal/zmosh"
)

// CounterName generates a session name like "dirname" or "dirname-N".
// Matches the zsh script's _zmosh_pick_name_counter logic.
func CounterName(dir string, existing []zmosh.Session) string {
	base := filepath.Base(dir)
	names := make(map[string]bool)
	for _, s := range existing {
		names[s.Name] = true
	}

	// First try: bare name
	if !names[base] {
		return base
	}

	// Increment: base-2, base-3, ...
	for n := 2; ; n++ {
		candidate := fmt.Sprintf("%s-%d", base, n)
		if !names[candidate] {
			return candidate
		}
	}
}

// DateName generates a session name like "dirname-MMDD".
// Matches the zsh script's _zmosh_pick_name_date logic.
func DateName(dir string) string {
	base := filepath.Base(dir)
	return fmt.Sprintf("%s-%s", base, time.Now().Format("0102"))
}

// DirBaseName extracts the base directory name from a path.
func DirBaseName(path string) string {
	return filepath.Base(path)
}
```

**Step 3: Run tests, commit**

```bash
go test ./internal/picker/ -v
git add internal/picker/
git commit -m "feat: session naming — counter and date modes"
```

---

## Phase 3: Dependency Checking

### Task 5: Dependency Checker

**Files:**
- Create: `internal/check/deps.go`
- Create: `internal/check/deps_test.go`

**Step 1: Write tests**

```go
// internal/check/deps_test.go
package check

import (
	"testing"
)

func TestCheckResultJSON(t *testing.T) {
	result := Result{
		Zmosh:  DepStatus{Installed: true, Version: "0.4.2", Path: "/opt/homebrew/bin/zmosh"},
		Zoxide: DepStatus{Installed: false},
		Fzf:    DepStatus{Installed: false},
		Shell:  "zsh",
		OS:     "darwin",
		Arch:   "arm64",
	}
	j, err := result.JSON()
	if err != nil {
		t.Fatal(err)
	}
	if len(j) == 0 {
		t.Error("expected non-empty JSON")
	}
}
```

**Step 2: Implement checker**

```go
// internal/check/deps.go
package check

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

type DepStatus struct {
	Installed bool   `json:"installed"`
	Version   string `json:"version,omitempty"`
	Path      string `json:"path,omitempty"`
}

type Result struct {
	Zmosh  DepStatus `json:"zmosh"`
	Zoxide DepStatus `json:"zoxide"`
	Fzf    DepStatus `json:"fzf"`
	Shell  string    `json:"shell"`
	OS     string    `json:"os"`
	Arch   string    `json:"arch"`
}

func (r Result) JSON() (string, error) {
	b, err := json.MarshalIndent(r, "", "  ")
	return string(b), err
}

// Run checks all dependencies and returns the result.
func Run() Result {
	r := Result{
		Shell: detectShell(),
		OS:    runtime.GOOS,
		Arch:  runtime.GOARCH,
	}
	r.Zmosh = checkDep("zmosh", "version")
	r.Zoxide = checkDep("zoxide", "--version")
	r.Fzf = checkDep("fzf", "--version")
	return r
}

func checkDep(name, versionFlag string) DepStatus {
	path, err := exec.LookPath(name)
	if err != nil {
		return DepStatus{Installed: false}
	}
	status := DepStatus{Installed: true, Path: path}
	if out, err := exec.Command(name, versionFlag).Output(); err == nil {
		status.Version = strings.TrimSpace(string(out))
	}
	return status
}

func detectShell() string {
	if shell, err := exec.Command("basename", "$SHELL").Output(); err == nil {
		return strings.TrimSpace(string(shell))
	}
	return "unknown"
}

// PrintHuman prints the check result in a human-readable format.
func (r Result) PrintHuman() {
	printDep("zmosh", r.Zmosh, true)
	printDep("zoxide", r.Zoxide, false)
	printDep("fzf", r.Fzf, false)
	fmt.Printf("\nPlatform: %s/%s, Shell: %s\n", r.OS, r.Arch, r.Shell)
}

func printDep(name string, d DepStatus, required bool) {
	status := "✓"
	if !d.Installed {
		if required {
			status = "✗"
		} else {
			status = "○"
		}
	}
	label := ""
	if required {
		label = " (required)"
	} else {
		label = " (optional)"
	}
	if d.Installed {
		fmt.Printf("  %s %s%s — %s\n", status, name, label, d.Version)
	} else {
		fmt.Printf("  %s %s%s — not found\n", status, name, label)
	}
}
```

**Step 3: Run tests, commit**

```bash
go test ./internal/check/ -v
git add internal/check/
git commit -m "feat: dependency checker with JSON and human output"
```

---

## Phase 4: Interactive Picker

### Task 6: Key Mapping

**Files:**
- Create: `internal/picker/keys.go`
- Create: `internal/picker/keys_test.go`

Port the zsh key map: `123456789abdefghijlmnopqrstuvwxy` (32 keys for up to 32 sessions).

**Step 1: Write tests**

```go
// internal/picker/keys_test.go
package picker

import "testing"

func TestKeyForIndex(t *testing.T) {
	tests := []struct {
		index int
		key   byte
	}{
		{0, '1'}, {1, '2'}, {8, '9'},
		{9, 'a'}, {10, 'b'}, {11, 'd'}, // skip 'c' (used for custom)
	}
	for _, tt := range tests {
		got := KeyForIndex(tt.index)
		if got != tt.key {
			t.Errorf("index %d: expected '%c', got '%c'", tt.index, tt.key, got)
		}
	}
}

func TestIndexForKey(t *testing.T) {
	idx, ok := IndexForKey('1')
	if !ok || idx != 0 {
		t.Errorf("expected 0, got %d (ok=%v)", idx, ok)
	}
	idx, ok = IndexForKey('a')
	if !ok || idx != 9 {
		t.Errorf("expected 9, got %d (ok=%v)", idx, ok)
	}
	_, ok = IndexForKey('c') // reserved
	if ok {
		t.Error("'c' should not be a valid session key")
	}
}
```

**Step 2: Implement keys**

```go
// internal/picker/keys.go
package picker

// keyChars maps session indices to keypress characters.
// Matches the zsh script: 123456789abdefghijlmnopqrstuvwxy
// Note: 'c' is reserved for custom name, 'k' is reserved for kill mode.
var keyChars = []byte("123456789abdefghijlmnopqrstuvwxy")

// KeyForIndex returns the key character for a session index.
func KeyForIndex(index int) byte {
	if index < 0 || index >= len(keyChars) {
		return '?'
	}
	return keyChars[index]
}

// IndexForKey returns the session index for a key character.
func IndexForKey(key byte) (int, bool) {
	for i, k := range keyChars {
		if k == key {
			return i, true
		}
	}
	return -1, false
}

// MaxSessions is the maximum number of sessions the picker can display.
const MaxSessions = len(keyChars)
```

**Step 3: Run tests, commit**

```bash
go test ./internal/picker/ -v
git add internal/picker/keys.go internal/picker/keys_test.go
git commit -m "feat: key mapping for session indices (1-9, a-y)"
```

---

### Task 7: Interactive TUI Picker

**Files:**
- Create: `internal/picker/picker.go`
- Modify: `cmd/zmosh-picker/main.go` (wire up stubs)

**Step 1: Implement the picker**

The picker uses raw terminal mode via `golang.org/x/term` for single-keypress input. It:
1. Fetches sessions from `zmosh list`
2. Renders the session list with key indicators
3. Reads a single keypress
4. Dispatches action (attach, new, zoxide, kill, etc.)
5. Loops after kill to show updated list

```go
// internal/picker/picker.go
package picker

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/nerveband/zmosh-picker/internal/zmosh"
	"golang.org/x/term"
)

// Run is the main interactive picker loop.
func Run() error {
	// Guard: skip if already in a zmosh session
	if os.Getenv("ZMX_SESSION") != "" && os.Getenv("ZPICK") == "" {
		return nil
	}
	// Guard: skip if not interactive
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return nil
	}
	// Guard: skip if zmosh not found
	if _, err := exec.LookPath("zmosh"); err != nil {
		return nil
	}

	for {
		sessions, err := zmosh.List()
		if err != nil {
			return fmt.Errorf("failed to list sessions: %w", err)
		}

		action, err := showPicker(sessions)
		if err != nil {
			return err
		}

		switch action.Type {
		case ActionAttach:
			return zmosh.Attach(action.Name)
		case ActionNew:
			cwd, _ := os.Getwd()
			name := CounterName(cwd, sessions)
			return zmosh.Attach(name)
		case ActionNewDate:
			cwd, _ := os.Getwd()
			name := DateName(cwd)
			return zmosh.Attach(name)
		case ActionCustom:
			name, err := promptCustomName()
			if err != nil || name == "" {
				continue // back to picker
			}
			return zmosh.Attach(name)
		case ActionZoxide:
			dir, err := runZoxide()
			if err != nil || dir == "" {
				continue
			}
			name := CounterName(dir, sessions)
			return zmosh.AttachInDir(name, dir)
		case ActionKill:
			if err := confirmAndKill(action.Name, sessions); err != nil {
				fmt.Fprintf(os.Stderr, "kill failed: %v\n", err)
			}
			continue // loop back to show updated list
		case ActionEscape:
			return nil // drop to plain shell
		}
	}
}

type ActionType int

const (
	ActionAttach ActionType = iota
	ActionNew
	ActionNewDate
	ActionCustom
	ActionZoxide
	ActionKill
	ActionEscape
)

type Action struct {
	Type ActionType
	Name string
}

func showPicker(sessions []zmosh.Session) (Action, error) {
	// Clear and render
	fmt.Print("\033[2J\033[H") // clear screen

	if len(sessions) > 0 {
		fmt.Printf("  zmosh: %d active sessions\n", len(sessions))
		for i, s := range sessions {
			if i >= MaxSessions {
				break
			}
			indicator := "."
			if s.Active {
				indicator = "*"
			}
			// Truncate path for display
			dir := truncatePath(s.StartedIn, 40)
			fmt.Printf("    %c) %s%s  (%d clients)  %s\n",
				KeyForIndex(i), indicator, s.Name, s.Clients, dir)
		}
		fmt.Println()
	}

	// Show actions
	cwd, _ := os.Getwd()
	defaultName := CounterName(cwd, sessions)
	fmt.Printf("  [Enter] new here: %s\n", defaultName)
	fmt.Println("  [z] pick dir first   [d] new +date")
	fmt.Println("  [c] custom name      [k] kill mode")
	fmt.Println("  [Esc] plain shell")

	// Read single keypress
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return Action{}, fmt.Errorf("failed to set raw mode: %w", err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	buf := make([]byte, 3)
	n, err := os.Stdin.Read(buf)
	if err != nil {
		return Action{}, err
	}

	key := buf[0]

	// Handle escape sequences
	if n == 1 && key == 27 {
		return Action{Type: ActionEscape}, nil
	}
	if n == 1 && (key == 13 || key == 10) { // Enter
		return Action{Type: ActionNew}, nil
	}

	switch key {
	case 'z':
		return Action{Type: ActionZoxide}, nil
	case 'd':
		return Action{Type: ActionNewDate}, nil
	case 'c':
		return Action{Type: ActionCustom}, nil
	case 'k':
		return enterKillMode(sessions)
	default:
		if idx, ok := IndexForKey(key); ok && idx < len(sessions) {
			return Action{Type: ActionAttach, Name: sessions[idx].Name}, nil
		}
	}

	return Action{Type: ActionEscape}, nil
}

func enterKillMode(sessions []zmosh.Session) (Action, error) {
	fmt.Print("\n  Kill which session? Press key or [Esc] to cancel: ")

	buf := make([]byte, 3)
	n, _ := os.Stdin.Read(buf)
	if n == 1 && buf[0] == 27 {
		return Action{Type: ActionEscape}, nil // they'll re-enter the loop
	}

	if idx, ok := IndexForKey(buf[0]); ok && idx < len(sessions) {
		return Action{Type: ActionKill, Name: sessions[idx].Name}, nil
	}

	return Action{Type: ActionEscape}, nil
}

func confirmAndKill(name string, sessions []zmosh.Session) error {
	if os.Getenv("ZMOSH_PICKER_NO_CONFIRM") == "1" {
		return zmosh.Kill(name)
	}
	fmt.Printf("\n  Kill \"%s\"? [y/N] ", name)
	buf := make([]byte, 1)
	os.Stdin.Read(buf)
	if buf[0] == 'y' || buf[0] == 'Y' {
		return zmosh.Kill(name)
	}
	return nil
}

func promptCustomName() (string, error) {
	// Temporarily restore terminal to cooked mode for line input
	fmt.Print("\n  Session name: ")
	var name string
	fmt.Scanln(&name)
	return strings.TrimSpace(name), nil
}

func runZoxide() (string, error) {
	cmd := exec.Command("zoxide", "query", "-i")
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func truncatePath(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}
	home := os.Getenv("HOME")
	if home != "" {
		path = strings.Replace(path, home, "~", 1)
	}
	if len(path) <= maxLen {
		return path
	}
	return "..." + path[len(path)-maxLen+3:]
}
```

**Step 2: Add x/term dependency**

```bash
go get golang.org/x/term
```

**Step 3: Wire up main.go stubs**

Replace the stub functions in `cmd/zmosh-picker/main.go` with real implementations that call into `internal/` packages.

**Step 4: Build and test interactively**

```bash
go build -o zmosh-picker ./cmd/zmosh-picker
./zmosh-picker
./zmosh-picker list
./zmosh-picker list --json
./zmosh-picker check
./zmosh-picker check --json
```

**Step 5: Commit**

```bash
git add .
git commit -m "feat: interactive TUI picker with raw terminal input"
```

---

## Phase 5: Shell Hook & Distribution

### Task 8: Shell Hook Installation

**Files:**
- Create: `internal/hook/install.go`
- Create: `internal/hook/zsh.go`
- Create: `internal/hook/bash.go`
- Create: `internal/hook/install_test.go`

The `install-hook` command detects the user's shell and adds the appropriate hook line. For zsh, it handles Powerlevel10k placement (hook must go before instant prompt).

**Step 1: Implement and test**

The hook line for all shells:
```sh
# zmosh-picker: session launcher
[[ -z "$ZMX_SESSION" ]] && command -v zmosh-picker &>/dev/null && zmosh-picker
```

Test: verify the hook line is generated correctly, verify p10k detection, verify idempotency (don't add twice).

**Step 2: Commit**

```bash
git add internal/hook/
git commit -m "feat: shell hook installation for zsh and bash"
```

---

### Task 9: Wire All Subcommands + Build

**Files:**
- Modify: `cmd/zmosh-picker/main.go`
- Create: `cmd/zmosh-picker/list.go`
- Create: `cmd/zmosh-picker/check_cmd.go`
- Create: `cmd/zmosh-picker/attach.go`
- Create: `cmd/zmosh-picker/kill_cmd.go`
- Create: `cmd/zmosh-picker/hook.go`

Split each subcommand into its own file in `cmd/zmosh-picker/` for clarity. Each calls into the corresponding `internal/` package.

Handle `--json` flag for `list` and `check` commands.

**Step 1: Implement all subcommand files**

**Step 2: Full integration test**

```bash
go build -o zmosh-picker ./cmd/zmosh-picker
./zmosh-picker version
./zmosh-picker check
./zmosh-picker check --json
./zmosh-picker list --json
./zmosh-picker --help
```

**Step 3: Commit**

```bash
git add cmd/
git commit -m "feat: wire all subcommands — list, check, attach, kill, install-hook"
```

---

### Task 10: Makefile + Goreleaser + Updated README

**Files:**
- Create: `Makefile`
- Create: `.goreleaser.yml`
- Modify: `README.md`
- Modify: `install.sh`

**Step 1: Create Makefile**

```makefile
VERSION ?= $(shell git describe --tags --always --dirty)
LDFLAGS = -ldflags "-X main.version=$(VERSION)"

build:
	go build $(LDFLAGS) -o zmosh-picker ./cmd/zmosh-picker

test:
	go test ./... -v

install: build
	cp zmosh-picker $(HOME)/.local/bin/

clean:
	rm -f zmosh-picker
```

**Step 2: Create .goreleaser.yml**

Configure for darwin/arm64, darwin/amd64, linux/arm64, linux/amd64.

**Step 3: Update README**

Document the new CLI interface, installation via `go install` and `brew`, all subcommands, and the `--json` output format.

**Step 4: Update install.sh**

Download pre-built binary from GitHub releases, fall back to `go install`.

**Step 5: Commit**

```bash
git add Makefile .goreleaser.yml README.md install.sh
git commit -m "feat: Makefile, goreleaser, updated README and install script"
```

---

### Task 11: Remove Old Zsh Script + Final Cleanup

**Files:**
- Remove: `zmosh-picker.zsh` (the renamed old script, kept for reference)
- Remove: `uninstall.sh` (replaced by `zmosh-picker install-hook --remove`)
- Run: `go vet ./...` and `go test ./...`

**Step 1: Final test suite**

```bash
go test ./... -v -count=1
go vet ./...
```

**Step 2: Clean up**

```bash
git rm zmosh-picker.zsh uninstall.sh
git add .
git commit -m "chore: remove old zsh script, final cleanup for Go v2"
```

---

## Summary

| Phase | Tasks | What it delivers |
|-------|-------|-----------------|
| 1. Setup | 1 | Go module, CLI routing, version command |
| 2. Core | 2-4 | zmosh parser, session ops, naming logic |
| 3. Deps | 5 | Dependency checker with JSON output |
| 4. Picker | 6-7 | Interactive TUI with single-keypress UX |
| 5. Distribution | 8-11 | Shell hooks, subcommands, goreleaser, README |

**Total: 11 tasks across 5 phases.**

After this plan completes, the zsync iOS app implementation plan can begin — it consumes `zmosh-picker list --json` and `zmosh-picker check --json`.
