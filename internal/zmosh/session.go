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
	// Try to get zmosh version (first line, last field: "zmosh\t\t0.4.0")
	if v, err := exec.Command("zmosh", "version").Output(); err == nil {
		ver := strings.TrimSpace(string(v))
		if idx := strings.IndexByte(ver, '\n'); idx >= 0 {
			ver = strings.TrimSpace(ver[:idx])
		}
		fields := strings.Fields(ver)
		if len(fields) >= 2 {
			ver = fields[len(fields)-1]
		}
		result.ZmoshVersion = ver
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
// Output is suppressed; the picker displays its own status message.
func Kill(name string) error {
	return exec.Command("zmosh", "kill", name).Run()
}
