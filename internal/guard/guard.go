package guard

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/nerveband/zpick/internal/picker"
	"github.com/nerveband/zpick/internal/zmosh"
	"golang.org/x/term"
)

const (
	reset   = "\033[0m"
	dim     = "\033[2m"
	boldYel = "\033[1;33m"
	boldWht = "\033[1;97m"
	boldGrn = "\033[1;32m"

	timeout = 10 * time.Second
)

// Run shows the guard prompt and returns a shell command to eval, or empty string.
// argv is the original command+args the user typed (e.g. ["claude", "--model", "opus"]).
func Run(argv []string) (string, error) {
	// Already in a zmosh session — exit silently
	if os.Getenv("ZMX_SESSION") != "" {
		return "", nil
	}

	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return "", nil // can't open tty, let the command run
	}
	defer tty.Close()

	// Show static prompt with timeout
	fmt.Fprintf(tty, "\n  %s⚡%s Not in a zmosh session. Press %sENTER%s to pick one (%ds)  %sesc%s %sskip%s\n",
		boldYel, reset, boldWht, reset, int(timeout.Seconds()), dim, reset, dim, reset)
	fmt.Fprintf(tty, "  %s>%s ", boldYel, reset)

	action := waitForKey(tty, timeout)

	// Clear the prompt line
	fmt.Fprintln(tty)

	switch action {
	case keyEnter:
		return runPicker(tty, argv)
	default:
		// ESC, timeout, Ctrl-C, or any other key — exit silently
		return "", nil
	}
}

type keyAction int

const (
	keyTimeout keyAction = iota
	keyEnter
	keyEscape
	keyOther
)

// waitForKey waits for a keypress or timeout. Uses raw mode on /dev/tty.
func waitForKey(tty *os.File, d time.Duration) keyAction {
	oldState, err := term.MakeRaw(int(tty.Fd()))
	if err != nil {
		return keyOther
	}
	defer term.Restore(int(tty.Fd()), oldState)

	resultCh := make(chan keyAction, 1)
	go func() {
		buf := make([]byte, 3)
		n, err := tty.Read(buf)
		if err != nil {
			resultCh <- keyOther
			return
		}
		if n >= 1 {
			switch buf[0] {
			case 13, 10: // Enter
				resultCh <- keyEnter
			case 27: // Escape
				resultCh <- keyEscape
			case 3: // Ctrl-C
				resultCh <- keyEscape
			default:
				resultCh <- keyOther
			}
		}
	}()

	select {
	case result := <-resultCh:
		return result
	case <-time.After(d):
		return keyTimeout
	}
}

// runPicker launches the full picker and returns the appropriate shell command.
// If the user picks a session, it includes the auto-launch env var for the original command.
func runPicker(tty *os.File, argv []string) (string, error) {
	cmd, err := picker.Run()
	if err != nil {
		return "", err
	}
	if cmd == "" {
		return "", nil // user escaped
	}

	// The picker returns commands like:
	//   exec zmosh attach "session-name"
	//   cd "/path" && exec zmosh attach "session-name"
	//
	// For new sessions (exec zmosh attach), we prepend ZPICK_AUTORUN
	// so the original command auto-launches inside the session.
	// For existing sessions, zmosh reconnects to a running shell,
	// so we can't auto-run — just print guidance.
	if len(argv) > 0 {
		encoded := encodeArgv(argv)
		if encoded != "" {
			// Check if this is an existing session by looking at the command
			// Existing sessions also use "exec zmosh attach", so we always set the env var.
			// It will only be picked up if a new shell starts (new session).
			// For existing sessions, also print guidance to tty.
			fmt.Fprintf(tty, "  %srun:%s %s\n", dim, reset, formatArgv(argv))
			return fmt.Sprintf("ZPICK_AUTORUN=%s %s", encoded, cmd), nil
		}
	}

	return cmd, nil
}

// encodeArgv encodes a command argv as base64 JSON for ZPICK_AUTORUN.
func encodeArgv(argv []string) string {
	if len(argv) == 0 {
		return ""
	}
	data, err := json.Marshal(argv)
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(data)
}

// DecodeArgv decodes a ZPICK_AUTORUN value back into argv.
func DecodeArgv(encoded string) ([]string, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("invalid base64: %w", err)
	}
	var argv []string
	if err := json.Unmarshal(data, &argv); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	if len(argv) == 0 {
		return nil, fmt.Errorf("empty argv")
	}
	return argv, nil
}

// Autorun reads ZPICK_AUTORUN, decodes argv, and execs the command.
// Called from the shell hook on session start.
func Autorun() error {
	encoded := os.Getenv("ZPICK_AUTORUN")
	if encoded == "" {
		return nil
	}

	argv, err := DecodeArgv(encoded)
	if err != nil {
		return nil // silently ignore decode errors
	}

	// Find the command in PATH
	path, err := exec.LookPath(argv[0])
	if err != nil {
		return fmt.Errorf("%s: command not found", argv[0])
	}

	// Unset ZPICK_AUTORUN before exec to prevent recursion
	os.Unsetenv("ZPICK_AUTORUN")

	return zmosh.ExecCommand(path, argv)
}

// formatArgv joins argv into a human-readable command string.
func formatArgv(argv []string) string {
	if len(argv) == 0 {
		return ""
	}
	result := argv[0]
	for _, a := range argv[1:] {
		result += " " + a
	}
	return result
}
