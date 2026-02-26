# Key Mode (Letters/Numbers) Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Let users toggle session picker keys between numbers-first (`1-9,a-y`) and letters-first (`a-y,1-9`) via config file and help screen.

**Architecture:** Add `ReadKeyMode`/`SetKeyMode` to the backend config package (same pattern as `SetBackend`). In `keys.go`, add a `LoadKeyMode()` that reads the config and swaps `keyChars`. Add `l` key to help screen to toggle and persist.

**Tech Stack:** Go, no new dependencies

---

### Task 1: Add key mode config read/write to backend package

**Files:**
- Modify: `internal/backend/config.go`
- Create: `internal/backend/config_test.go` (or add to existing)

**Step 1: Write the failing test**

Add to `internal/backend/config_test.go`:

```go
package backend

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadKeyMode_Default(t *testing.T) {
	// Use a temp dir so no config file exists
	old := os.Getenv("XDG_CONFIG_HOME")
	tmp := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", tmp)
	defer os.Setenv("XDG_CONFIG_HOME", old)

	mode := ReadKeyMode()
	if mode != "numbers" {
		t.Errorf("expected default 'numbers', got %q", mode)
	}
}

func TestSetAndReadKeyMode(t *testing.T) {
	old := os.Getenv("XDG_CONFIG_HOME")
	tmp := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", tmp)
	defer os.Setenv("XDG_CONFIG_HOME", old)

	if err := SetKeyMode("letters"); err != nil {
		t.Fatalf("SetKeyMode: %v", err)
	}

	mode := ReadKeyMode()
	if mode != "letters" {
		t.Errorf("expected 'letters', got %q", mode)
	}

	// Verify file contents
	data, _ := os.ReadFile(filepath.Join(tmp, "zpick", "keys"))
	if got := string(data); got != "letters\n" {
		t.Errorf("file contents: %q", got)
	}
}

func TestSetKeyMode_Invalid(t *testing.T) {
	err := SetKeyMode("emoji")
	if err == nil {
		t.Error("expected error for invalid key mode")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/backend/ -run TestReadKeyMode -v && go test ./internal/backend/ -run TestSetAndReadKeyMode -v && go test ./internal/backend/ -run TestSetKeyMode_Invalid -v`
Expected: FAIL — `ReadKeyMode` and `SetKeyMode` undefined

**Step 3: Write minimal implementation**

Add to `internal/backend/config.go`, after the `ReadUDP`/`SetUDP` block:

```go
// ReadKeyMode returns the configured key mode ("numbers" or "letters").
// Defaults to "numbers" if not configured.
func ReadKeyMode() string {
	data, err := os.ReadFile(filepath.Join(ConfigDir(), "keys"))
	if err != nil {
		return "numbers"
	}
	mode := strings.TrimSpace(string(data))
	if mode == "letters" {
		return "letters"
	}
	return "numbers"
}

// SetKeyMode writes the key mode to the config file.
func SetKeyMode(mode string) error {
	if mode != "numbers" && mode != "letters" {
		return fmt.Errorf("invalid key mode %q (valid: numbers, letters)", mode)
	}
	dir := ConfigDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "keys"), []byte(mode+"\n"), 0644)
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/backend/ -run "TestReadKeyMode|TestSetAndReadKeyMode|TestSetKeyMode_Invalid" -v`
Expected: PASS (3 tests)

**Step 5: Commit**

```bash
git add internal/backend/config.go internal/backend/config_test.go
git commit -m "feat: add ReadKeyMode/SetKeyMode for letters/numbers config"
```

---

### Task 2: Add LoadKeyMode to keys.go to swap keyChars

**Files:**
- Modify: `internal/picker/keys.go`
- Modify: `internal/picker/keys_test.go`

**Step 1: Write the failing tests**

Add to `internal/picker/keys_test.go`:

```go
func TestLoadKeyMode_Numbers(t *testing.T) {
	LoadKeyMode("numbers")
	if KeyForIndex(0) != '1' {
		t.Errorf("numbers mode: index 0 should be '1', got '%c'", KeyForIndex(0))
	}
	if KeyForIndex(9) != 'a' {
		t.Errorf("numbers mode: index 9 should be 'a', got '%c'", KeyForIndex(9))
	}
}

func TestLoadKeyMode_Letters(t *testing.T) {
	LoadKeyMode("letters")
	defer LoadKeyMode("numbers") // restore for other tests

	if KeyForIndex(0) != 'a' {
		t.Errorf("letters mode: index 0 should be 'a', got '%c'", KeyForIndex(0))
	}
	// 'c' is skipped, so index 2 should be 'd'
	if KeyForIndex(2) != 'd' {
		t.Errorf("letters mode: index 2 should be 'd', got '%c'", KeyForIndex(2))
	}
	// Numbers should come after letters
	_, ok := IndexForKey('1')
	if !ok {
		t.Error("letters mode: '1' should still be a valid key")
	}
}

func TestLoadKeyMode_LettersSkipsReserved(t *testing.T) {
	LoadKeyMode("letters")
	defer LoadKeyMode("numbers")

	_, ok := IndexForKey('c')
	if ok {
		t.Error("'c' should still be reserved in letters mode")
	}
	_, ok = IndexForKey('k')
	if ok {
		t.Error("'k' should still be reserved in letters mode")
	}
}

func TestLoadKeyMode_LettersMaxSessions(t *testing.T) {
	LoadKeyMode("letters")
	defer LoadKeyMode("numbers")

	// Both modes should have the same number of keys
	if MaxSessions != 32 {
		t.Errorf("expected 32 max sessions in letters mode, got %d", MaxSessions)
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/picker/ -run "TestLoadKeyMode" -v`
Expected: FAIL — `LoadKeyMode` undefined

**Step 3: Write minimal implementation**

Replace the contents of `internal/picker/keys.go` with:

```go
package picker

// Key character sequences for each mode.
// Both skip 'c' (custom) and 'k' (kill).
const (
	numbersFirst = "123456789abdefghijlmnopqrstuvwxy"
	lettersFirst = "abdefghijlmnopqrstuvwxy123456789"
)

// keyChars maps session indices to keypress characters.
var keyChars = []byte(numbersFirst)

// MaxSessions is the maximum number of sessions the picker can display.
var MaxSessions = len(keyChars)

// LoadKeyMode sets the active key mapping. mode is "numbers" or "letters".
func LoadKeyMode(mode string) {
	if mode == "letters" {
		keyChars = []byte(lettersFirst)
	} else {
		keyChars = []byte(numbersFirst)
	}
	MaxSessions = len(keyChars)
}

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
```

**Step 4: Run all picker tests to verify they pass**

Run: `go test ./internal/picker/ -v`
Expected: ALL PASS (existing tests pass because default is numbers-first, new tests pass too)

**Step 5: Commit**

```bash
git add internal/picker/keys.go internal/picker/keys_test.go
git commit -m "feat: add LoadKeyMode to swap between letters-first and numbers-first"
```

---

### Task 3: Call LoadKeyMode at picker startup

**Files:**
- Modify: `internal/picker/picker.go` (the `Run` function)

**Step 1: No new test needed** — this is wiring. The existing tests cover `LoadKeyMode` behavior.

**Step 2: Add the call**

In `internal/picker/picker.go`, at the start of the `Run` function (before the picker loop), add:

```go
// Load key mode preference (letters-first or numbers-first)
LoadKeyMode(backend.ReadKeyMode())
```

Add `"github.com/nerveband/zpick/internal/backend"` to imports if not already present (it should already be there).

**Step 3: Verify build and tests pass**

Run: `go build ./... && go test ./... -count=1`
Expected: Build succeeds, all tests pass

**Step 4: Commit**

```bash
git add internal/picker/picker.go
git commit -m "feat: load key mode config at picker startup"
```

---

### Task 4: Add key mode toggle to help screen

**Files:**
- Modify: `internal/picker/help.go`

**Step 1: No separate test file** — the help screen is interactive TUI rendering. We verify by building and manual check.

**Step 2: Add the toggle key and display**

In `internal/picker/help.go`:

**a)** In `renderHelp()`, after the UDP section (line ~96) and before the blank line, add a keys-mode line:

```go
	// Key mode
	keyMode := backend.ReadKeyMode()
	keyLabel := "1-9,a-y"
	if keyMode == "letters" {
		keyLabel = "a-y,1-9"
	}
	fmt.Fprintf(tty, "    %sl%s  keys       %s%s (%s)%s\n", magenta, reset, boldWht, keyMode, keyLabel, reset)
```

**b)** In `renderHelp()`, update the Keys section header (line 56) to be dynamic:

Replace:
```go
	fmt.Fprintf(tty, "    %s1-9,a-y%s  attach session       %senter%s  new session\n", boldYel, reset, boldGrn, reset)
```
With:
```go
	keyRange := "1-9,a-y"
	if backend.ReadKeyMode() == "letters" {
		keyRange = "a-y,1-9"
	}
	fmt.Fprintf(tty, "    %s%s%s  attach session       %senter%s  new session\n", boldYel, keyRange, reset, boldGrn, reset)
```

**c)** In `showHelpConfig()`, add the `'l'` case to the switch (after `'u'`):

```go
	case 'l':
		toggleKeyMode()
```

**d)** Add the `toggleKeyMode` function:

```go
func toggleKeyMode() {
	current := backend.ReadKeyMode()
	next := "letters"
	if current == "letters" {
		next = "numbers"
	}
	backend.SetKeyMode(next)
	LoadKeyMode(next)
}
```

**Step 3: Verify build passes**

Run: `go build ./...`
Expected: Build succeeds

**Step 4: Commit**

```bash
git add internal/picker/help.go
git commit -m "feat: add key mode toggle (l key) to help screen"
```

---

### Task 5: Update README and CHANGELOG

**Files:**
- Modify: `README.md`
- Modify: `CHANGELOG.md`

**Step 1: Update README**

In the Keys table, add a row for `l` (on help screen). Also update the `h` key description or add a note about the key mode setting.

Add after the Status indicators section:

```markdown
### Key mode

By default, sessions are labeled `1-9` then `a-y`. If you're on a mobile keyboard where letters are the default view, switch to letters-first mode:

Press `h` for the help screen, then `l` to toggle between `numbers` and `letters` mode. The setting is saved to `~/.config/zpick/keys`.
```

**Step 2: Update CHANGELOG**

The changelog entry will be added when we tag the release. No action needed now — goreleaser auto-generates it from commits.

**Step 3: Commit**

```bash
git add README.md
git commit -m "docs: add key mode section to README"
```
