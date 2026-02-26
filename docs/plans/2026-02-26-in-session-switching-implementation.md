# In-Session Switching Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Allow users to switch sessions from inside an existing session with 4 keystrokes, add `/usr/local/bin` symlink for mosh compatibility, and update the SVG demo.

**Architecture:** When `zp` detects `InSession()`, show the full picker. On session selection, write target to `~/.cache/zpick/switch-target`, output backend detach command. Shell hook in new shell reads the file and auto-attaches. Symlink created by `install-hook` and `make install` only (never auto-recreated by `upgrade`).

**Tech Stack:** Go, shell hooks (bash/zsh/fish), SVG

---

### Task 1: Add `DetachCommand()` to Backend Interface

**Files:**
- Modify: `internal/backend/types.go:18-35`
- Test: `internal/backend/tmux/tmux_test.go`
- Test: `internal/backend/zellij/zellij_test.go`
- Test: `internal/backend/zmosh/zmosh_test.go`
- Test: `internal/backend/zmx/zmx_test.go`
- Test: `internal/backend/shpool/shpool_test.go`

**Step 1: Write failing tests for DetachCommand on each backend**

Add to each backend's test file:

```go
// In tmux_test.go
func TestDetachCommand(t *testing.T) {
	b := New()
	cmd := b.DetachCommand()
	if cmd != "tmux detach-client" {
		t.Errorf("DetachCommand() = %q, want %q", cmd, "tmux detach-client")
	}
}

// In zellij_test.go
func TestDetachCommand(t *testing.T) {
	b := New()
	cmd := b.DetachCommand()
	if cmd != "zellij action detach" {
		t.Errorf("DetachCommand() = %q, want %q", cmd, "zellij action detach")
	}
}

// In zmosh_test.go
func TestDetachCommand(t *testing.T) {
	b := New()
	cmd := b.DetachCommand()
	if cmd != "zmx detach" {
		t.Errorf("DetachCommand() = %q, want %q", cmd, "zmx detach")
	}
}

// In zmx_test.go
func TestDetachCommand(t *testing.T) {
	b := New()
	cmd := b.DetachCommand()
	if cmd != "zmx detach" {
		t.Errorf("DetachCommand() = %q, want %q", cmd, "zmx detach")
	}
}

// In shpool_test.go
func TestDetachCommand(t *testing.T) {
	b := New()
	cmd := b.DetachCommand()
	if cmd != "shpool detach" {
		t.Errorf("DetachCommand() = %q, want %q", cmd, "shpool detach")
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/backend/... -v -run TestDetachCommand`
Expected: Compile error — `DetachCommand` not defined on interface or implementations.

**Step 3: Add `DetachCommand()` to interface and implement on all backends**

In `internal/backend/types.go`, add to the `Backend` interface:

```go
// DetachCommand returns the shell command to detach from the current session.
DetachCommand() string
```

In each backend implementation, add the method:

```go
// tmux/tmux.go
func (t *Tmux) DetachCommand() string { return "tmux detach-client" }

// zellij/zellij.go
func (z *Zellij) DetachCommand() string { return "zellij action detach" }

// zmosh/zmosh.go
func (z *Zmosh) DetachCommand() string { return "zmx detach" }

// zmx/zmx.go
func (z *Zmx) DetachCommand() string { return "zmx detach" }

// shpool/shpool.go
func (s *Shpool) DetachCommand() string { return "shpool detach" }
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/backend/... -v -run TestDetachCommand`
Expected: All PASS.

**Step 5: Commit**

```bash
git add internal/backend/types.go internal/backend/*/
git commit -m "feat: add DetachCommand() to Backend interface"
```

---

### Task 2: Create switch-target file handling (`internal/switcher`)

**Files:**
- Create: `internal/switcher/switcher.go`
- Create: `internal/switcher/switcher_test.go`

**Step 1: Write failing tests**

```go
package switcher

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWriteAndRead(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "switch-target")
	SetPath(path) // allow overriding for tests

	target := Target{Action: "attach", Name: "my-session"}
	if err := Write(target); err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	got, err := Read()
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}
	if got.Action != "attach" || got.Name != "my-session" {
		t.Errorf("Read() = %+v, want action=attach name=my-session", got)
	}
}

func TestReadDeletesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "switch-target")
	SetPath(path)

	Write(Target{Action: "attach", Name: "test"})
	Read()

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("Read() should delete the switch-target file")
	}
}

func TestReadStaleFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "switch-target")
	SetPath(path)

	Write(Target{Action: "attach", Name: "test"})
	// Backdate the file by 60 seconds
	old := time.Now().Add(-60 * time.Second)
	os.Chtimes(path, old, old)

	_, err := Read()
	if err == nil {
		t.Error("Read() should return error for stale file (>30s)")
	}
}

func TestReadMissingFile(t *testing.T) {
	dir := t.TempDir()
	SetPath(filepath.Join(dir, "nonexistent"))

	_, err := Read()
	if err == nil {
		t.Error("Read() should return error for missing file")
	}
}

func TestWriteWithDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "switch-target")
	SetPath(path)

	target := Target{Action: "new", Name: "scratch", Dir: "/tmp/work"}
	Write(target)

	got, _ := Read()
	if got.Dir != "/tmp/work" {
		t.Errorf("Dir = %q, want %q", got.Dir, "/tmp/work")
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/switcher/... -v`
Expected: Compile error — package doesn't exist.

**Step 3: Implement the switcher package**

```go
package switcher

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Target describes what session to switch to after detaching.
type Target struct {
	Action string `json:"action"` // "attach" or "new"
	Name   string `json:"name"`
	Dir    string `json:"dir,omitempty"`
}

var overridePath string

// SetPath overrides the switch-target file path (for testing).
func SetPath(p string) { overridePath = p }

func path() string {
	if overridePath != "" {
		return overridePath
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cache", "zpick", "switch-target")
}

// Write saves the switch target to disk.
func Write(t Target) error {
	p := path()
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return err
	}
	data, err := json.Marshal(t)
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0644)
}

// Read reads and deletes the switch target file.
// Returns error if file is missing or stale (>30s old).
func Read() (Target, error) {
	p := path()
	info, err := os.Stat(p)
	if err != nil {
		return Target{}, fmt.Errorf("no switch target")
	}
	if time.Since(info.ModTime()) > 30*time.Second {
		os.Remove(p)
		return Target{}, fmt.Errorf("stale switch target (>30s)")
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return Target{}, err
	}
	os.Remove(p)
	var t Target
	if err := json.Unmarshal(data, &t); err != nil {
		return Target{}, err
	}
	return t, nil
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/switcher/... -v`
Expected: All PASS.

**Step 5: Commit**

```bash
git add internal/switcher/
git commit -m "feat: add switcher package for switch-target file handling"
```

---

### Task 3: Add `zp resume` subcommand

**Files:**
- Create: `cmd/zp/resume.go`
- Modify: `cmd/zp/main.go:34-91` (add case for "resume")

**Step 1: Create `resume.go`**

```go
package main

import (
	"fmt"

	"github.com/nerveband/zpick/internal/switcher"
)

func runResume() error {
	b, err := loadBackend(false)
	if err != nil {
		return err
	}

	target, err := switcher.Read()
	if err != nil {
		// No switch target — nothing to do, not an error
		return nil
	}

	switch target.Action {
	case "attach":
		fmt.Print("exec " + b.AttachCommand(target.Name, ""))
	case "new":
		if target.Dir != "" {
			fmt.Printf("cd %q && exec %s", target.Dir, b.AttachCommand(target.Name, ""))
		} else {
			fmt.Print("exec " + b.AttachCommand(target.Name, ""))
		}
	}
	return nil
}
```

**Step 2: Register the subcommand in `main.go`**

Add this case in the switch block in `main.go`, after the "autorun" case:

```go
case "resume":
	if err := runResume(); err != nil {
		fmt.Fprintf(os.Stderr, "zp: %v\n", err)
		os.Exit(1)
	}
```

**Step 3: Build and verify**

Run: `go build ./cmd/zp && ./zp resume`
Expected: No output (no switch-target file exists), exit 0.

**Step 4: Commit**

```bash
git add cmd/zp/resume.go cmd/zp/main.go
git commit -m "feat: add 'zp resume' subcommand for switch-target handoff"
```

---

### Task 4: Modify picker for in-session mode

**Files:**
- Modify: `internal/picker/picker.go:49-62` (replace InSession guard)
- Modify: `internal/picker/picker.go:140-229` (showPicker display)

**Step 1: Replace the InSession guard with in-session picker mode**

In `picker.go`, replace the guard block (lines 52-62) with logic that sets an `inSession` flag and the current session name. Then, in the action switch block, when `inSession` is true, write the switch target and return the detach command instead of the attach command.

Replace the guard block at the top of `Run()`:

```go
// Detect in-session mode
inSession := b.InSession() && os.Getenv("ZPICK") == ""
var currentSession string
if inSession {
	currentSession = os.Getenv(b.SessionEnvVar())
}
```

**Step 2: Update showPicker to accept and display current session**

Add `currentSession string` parameter to `showPicker`. In the header, if `currentSession != ""`, show `(in: name <-)`. In the session list, mark the current session with `<-` instead of `*`/`.`.

Change the `showPicker` signature:

```go
func showPicker(tty *os.File, b backend.Backend, sessions []backend.Session, currentSession string) (Action, error) {
```

Update the header display:

```go
if currentSession != "" {
	fmt.Fprintf(tty, "  %s%s%s %s%d session%s%s  %s(in: %s ←)%s\n\n",
		boldCyan, b.Name(), reset, dim, len(sessions), plural, reset,
		dim, currentSession, reset)
} else {
	fmt.Fprintf(tty, "  %s%s%s %s%d session%s%s\n\n", ...)  // existing line
}
```

Update the session row display — if `s.Name == currentSession`, show `←` instead of `*`/`.`:

```go
indicator := fmt.Sprintf("%s.%s", dim, reset)
if s.Name == currentSession {
	indicator = fmt.Sprintf("%s←%s", boldCyan, reset)
} else if s.Active {
	indicator = fmt.Sprintf("%s*%s", boldGrn, reset)
}
```

Update all `showPicker` call sites to pass `currentSession`.

**Step 3: Update action handlers for in-session mode**

In the action switch block in `Run()`, wrap attach/new/custom/zoxide/date actions. When `inSession`, write the switch-target and return the detach command:

```go
case ActionAttach:
	if inSession {
		switcher.Write(switcher.Target{Action: "attach", Name: action.Name})
		return b.DetachCommand(), nil
	}
	return "exec " + b.AttachCommand(action.Name, ""), nil

case ActionNew:
	cwd, _ := os.Getwd()
	name := CounterName(cwd, sessions)
	fmt.Fprintf(tty, "\n  %s>%s %s%s%s\n\n", boldGrn, reset, boldWht, name, reset)
	if inSession {
		switcher.Write(switcher.Target{Action: "new", Name: name})
		return b.DetachCommand(), nil
	}
	return "exec " + b.AttachCommand(name, ""), nil
```

Apply the same pattern to `ActionNewDate`, `ActionCustom` (in `handleCustom`), and `ActionZoxide`.

Add import for `"github.com/nerveband/zpick/internal/switcher"` at the top of picker.go.

**Step 4: Build and verify compilation**

Run: `go build ./cmd/zp`
Expected: Compiles cleanly.

**Step 5: Run existing tests**

Run: `go test ./... -v`
Expected: All existing tests still pass.

**Step 6: Commit**

```bash
git add internal/picker/picker.go
git commit -m "feat: in-session picker with detach+switch-target handoff"
```

---

### Task 5: Update shell hooks for switch-target auto-resume

**Files:**
- Modify: `internal/hook/install.go:45-85` (GenerateHookBlock)
- Modify: `internal/hook/fish.go:34-73` (GenerateFishHookBlock)
- Modify: `internal/hook/install_test.go`

**Step 1: Write failing test for switch-target in hook block**

Add to `install_test.go`:

```go
func TestGenerateHookBlockContainsSwitchTarget(t *testing.T) {
	block := GenerateHookBlock([]string{"claude"})

	if !strings.Contains(block, "switch-target") {
		t.Error("block should contain switch-target check")
	}
	if !strings.Contains(block, "_zpick_switch") {
		t.Error("block should contain _zpick_switch function")
	}
	if !strings.Contains(block, "zp resume") {
		t.Error("block should reference zp resume command")
	}
}

func TestGenerateFishHookBlockContainsSwitchTarget(t *testing.T) {
	block := GenerateFishHookBlock([]string{"claude"})

	if !strings.Contains(block, "switch-target") {
		t.Error("fish block should contain switch-target check")
	}
	if !strings.Contains(block, "zp resume") {
		t.Error("fish block should reference zp resume command")
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/hook/... -v -run SwitchTarget`
Expected: FAIL — blocks don't contain switch-target references.

**Step 3: Add switch-target check to GenerateHookBlock (bash/zsh)**

In `GenerateHookBlock`, after the autorun block and before the guard function, add:

```go
// Switch-target: resume after in-session detach
b.WriteString("if [[ -f \"$HOME/.cache/zpick/switch-target\" ]]; then\n")
b.WriteString("  _zpick_switch() {\n")
b.WriteString("    precmd_functions=(${precmd_functions:#_zpick_switch})\n")
b.WriteString("    eval \"$(command zp resume)\"\n")
b.WriteString("  }\n")
b.WriteString("  precmd_functions+=(_zpick_switch)\n")
b.WriteString("fi\n")
```

**Step 4: Add switch-target check to GenerateFishHookBlock**

In `GenerateFishHookBlock`, after the autorun block, add:

```go
// Switch-target: resume after in-session detach
b.WriteString("if test -f \"$HOME/.cache/zpick/switch-target\"\n")
b.WriteString("  eval (command zp resume)\n")
b.WriteString("end\n")
```

**Step 5: Run tests to verify they pass**

Run: `go test ./internal/hook/... -v`
Expected: All PASS (including new and existing tests).

**Step 6: Commit**

```bash
git add internal/hook/
git commit -m "feat: add switch-target auto-resume to shell hooks"
```

---

### Task 6: Add PATH symlink to `install-hook`

**Files:**
- Create: `internal/hook/symlink.go`
- Modify: `internal/hook/install.go:88-102` (Install function)
- Modify: `cmd/zp/upgrade.go`

**Step 1: Create symlink helper**

```go
package hook

import (
	"fmt"
	"os"
	"path/filepath"
)

const symlinkPath = "/usr/local/bin/zp"

// InstallSymlink creates a symlink from /usr/local/bin/zp to the binary
// in ~/.local/bin/zp. Prints a sudo hint if permissions deny it.
func InstallSymlink() {
	home, _ := os.UserHomeDir()
	target := filepath.Join(home, ".local", "bin", "zp")

	// Check target exists
	if _, err := os.Stat(target); os.IsNotExist(err) {
		return
	}

	// Remove old symlink if it exists (may point elsewhere)
	os.Remove(symlinkPath)

	if err := os.Symlink(target, symlinkPath); err != nil {
		fmt.Printf("  note: run 'sudo ln -sf %s %s' for system-wide PATH\n", target, symlinkPath)
		return
	}
	fmt.Printf("  symlinked %s -> %s\n", symlinkPath, target)
}

// CheckSymlink prints a note if the symlink doesn't exist.
// Used by `zp upgrade` — never auto-creates.
func CheckSymlink() {
	if _, err := os.Lstat(symlinkPath); os.IsNotExist(err) {
		fmt.Println("  note: /usr/local/bin/zp not found — run 'zp install-hook' to add it")
	}
}
```

**Step 2: Call `InstallSymlink()` from `Install()`**

In `internal/hook/install.go`, at the end of the `Install()` function, add:

```go
func Install() error {
	shell := detectShell()
	// ... existing switch ...

	// After successful hook install, also set up PATH symlink
	InstallSymlink()
	return err  // return the error from the shell-specific install
}
```

Actually, restructure slightly so symlink runs after the shell install succeeds:

```go
func Install() error {
	shell := detectShell()
	var err error
	switch shell {
	case "zsh":
		err = installShell(zshrcPath())
	case "bash":
		err = installShell(bashrcPath())
	case "fish":
		err = installFish()
	default:
		apps, _ := guard.ReadConfig()
		block := GenerateHookBlock(apps)
		return fmt.Errorf("unsupported shell: %s\nManually add this to your shell config:\n\n%s", shell, block)
	}
	if err == nil {
		InstallSymlink()
	}
	return err
}
```

**Step 3: Call `CheckSymlink()` from upgrade**

In `cmd/zp/upgrade.go`, after a successful upgrade, call the check:

```go
func runUpgrade() error {
	err := update.Upgrade(version)
	if err == nil {
		hook.CheckSymlink()
	}
	return err
}
```

Add import for `"github.com/nerveband/zpick/internal/hook"`.

**Step 4: Build and verify**

Run: `go build ./cmd/zp`
Expected: Compiles cleanly.

**Step 5: Run all tests**

Run: `go test ./... -v`
Expected: All PASS.

**Step 6: Commit**

```bash
git add internal/hook/symlink.go internal/hook/install.go cmd/zp/upgrade.go
git commit -m "feat: add /usr/local/bin/zp symlink for mosh/system PATH"
```

---

### Task 7: Update SVG animation

**Files:**
- Modify: `assets/screenshot.svg`

**Step 1: Add 5th frame showing in-session picker**

Add a new `<g class="f5">` group after frame 4 (shpool) showing the zellij in-session view:

```
zellij 3 sessions  (in: frontend ←)

1  frontend  ← ~/projects/frontend
2  backend   .  ~/projects/backend
3  scratch   .  ~/tmp
```

**Step 2: Adjust keyframe timing from 4 frames to 5**

Update CSS animations: each frame now gets 1/5 of the cycle instead of 1/4.
- Total cycle: 20s (4s visible per frame)
- frame1: 0-20% visible, hidden rest
- frame2: 20-40% visible
- frame3: 40-60% visible
- frame4: 60-80% visible
- frame5: 80-100% visible

Add `.f5 { animation: frame5 20s ease-in-out infinite; }` and the `@keyframes frame5` rule.

**Step 3: Update tagline**

Change the bottom tagline from:
```
one picker for zmosh, tmux, zellij, and shpool
```
to:
```
switch sessions from anywhere — zmosh · tmux · zellij · shpool
```

**Step 4: Verify SVG renders correctly**

Open `assets/screenshot.svg` in a browser and verify all 5 frames cycle correctly with smooth transitions.

**Step 5: Commit**

```bash
git add assets/screenshot.svg
git commit -m "feat: update SVG animation with in-session switching frame"
```

---

### Task 8: Re-run install-hook and full test suite

**Step 1: Run full test suite**

Run: `go test ./... -v`
Expected: All PASS.

**Step 2: Build and install locally**

Run: `make install`
Expected: Binary built, signed, symlinked.

**Step 3: Re-install hook to pick up new switch-target block**

Run: `./zp install-hook`
Expected: Hook updated in shell config, symlink created/confirmed.

**Step 4: Manual smoke test**

1. Start a session: `zp` → pick or create
2. Inside session: run `zp` → should see picker with `←` on current session
3. Pick another session → should detach, auto-relaunch, attach to new session
4. Verify `mosh localhost -- zp` works (if mosh available)

**Step 5: Final commit**

```bash
git add -A
git commit -m "chore: final integration verification"
```
