package backend

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/term"
)

// InAnySession reports whether the current shell is already running inside a
// supported session backend. It treats stale inherited env vars as "not in a
// session" unless the process tree also indicates a live backend ancestor.
func InAnySession() bool {
	if os.Getenv("ZPICK_SESSION") != "" {
		return true
	}
	if os.Getenv("TMUX") != "" && tmuxSessionActive() {
		return true
	}
	if !hasSessionEnv() {
		return false
	}
	return processTreeHasSessionBackend(os.Getppid())
}

// ShouldAutostart reports whether zpick should launch the picker for the
// current shell startup.
func ShouldAutostart() bool {
	ttyReady := term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd()))
	if !ttyReady {
		ttyReady = hasControllingTTY()
	}
	return shouldAutostartFromState(ttyReady, InAnySession(), os.Getenv("TERM"))
}

func hasControllingTTY() bool {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return false
	}
	_ = tty.Close()
	return true
}

func shouldAutostartFromState(ttyReady, inSession bool, termName string) bool {
	if !ttyReady || inSession {
		return false
	}
	termName = strings.TrimSpace(termName)
	return termName != "" && termName != "dumb"
}

func hasSessionEnv() bool {
	for _, name := range append(AllSessionEnvVars(), "ZPICK_SESSION") {
		if os.Getenv(name) != "" {
			return true
		}
	}
	return false
}

func tmuxSessionActive() bool {
	cmd := exec.Command("tmux", "display-message", "-p", "#S")
	cmd.Env = os.Environ()
	return cmd.Run() == nil
}

func processTreeHasSessionBackend(pid int) bool {
	for depth := 0; pid > 1 && depth < 32; depth++ {
		command, err := processCommand(pid)
		if err == nil && commandHasSessionBackend(command) {
			return true
		}
		next, err := parentPID(pid)
		if err != nil || next == pid {
			return false
		}
		pid = next
	}
	return false
}

func parentPID(pid int) (int, error) {
	out, err := exec.Command("ps", "-o", "ppid=", "-p", strconv.Itoa(pid)).Output()
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(out)))
}

func processCommand(pid int) (string, error) {
	out, err := exec.Command("ps", "-o", "command=", "-p", strconv.Itoa(pid)).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func commandHasSessionBackend(command string) bool {
	if command == "" {
		return false
	}

	needles := map[string]struct{}{
		"tmux":   {},
		"zellij": {},
		"shpool": {},
		"zmosh":  {},
		"zmx":    {},
	}

	for _, field := range strings.Fields(strings.ToLower(command)) {
		trimmed := strings.Trim(field, "[]():")
		if trimmed == "" {
			continue
		}
		base := filepath.Base(trimmed)
		if _, ok := needles[base]; ok {
			return true
		}
	}

	return false
}
