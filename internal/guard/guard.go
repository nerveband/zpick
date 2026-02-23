package guard

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/nerveband/zpick/internal/backend"
	"github.com/nerveband/zpick/internal/picker"
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
func Run(b backend.Backend, argv []string) (string, error) {
	// Already in a session — exit silently
	if b.InSession() {
		return "", nil
	}

	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return "", nil
	}
	defer tty.Close()

	fmt.Fprintf(tty, "\n  %s⚡%s Not in a %s session. Press %sENTER%s to pick one (%ds)  %sesc%s %sskip%s\n",
		boldYel, reset, b.Name(), boldWht, reset, int(timeout.Seconds()), dim, reset, dim, reset)
	fmt.Fprintf(tty, "  %s>%s ", boldYel, reset)

	action := waitForKey(tty, timeout)

	fmt.Fprintln(tty)

	switch action {
	case keyEnter:
		return runPicker(tty, b, argv)
	default:
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
			case 13, 10:
				resultCh <- keyEnter
			case 27:
				resultCh <- keyEscape
			case 3:
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

func runPicker(tty *os.File, b backend.Backend, argv []string) (string, error) {
	cmd, err := picker.Run(b)
	if err != nil {
		return "", err
	}
	if cmd == "" {
		return "", nil
	}

	if len(argv) > 0 {
		encoded := encodeArgv(argv)
		if encoded != "" {
			fmt.Fprintf(tty, "  %srun:%s %s\n", dim, reset, formatArgv(argv))
			return fmt.Sprintf("ZPICK_AUTORUN=%s %s", encoded, cmd), nil
		}
	}

	return cmd, nil
}

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
func Autorun() error {
	encoded := os.Getenv("ZPICK_AUTORUN")
	if encoded == "" {
		return nil
	}

	argv, err := DecodeArgv(encoded)
	if err != nil {
		return nil
	}

	path, err := exec.LookPath(argv[0])
	if err != nil {
		return fmt.Errorf("%s: command not found", argv[0])
	}

	os.Unsetenv("ZPICK_AUTORUN")

	return backend.ExecCommand(path, argv)
}

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
