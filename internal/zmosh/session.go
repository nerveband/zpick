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

// resolveZmxDir finds the zmx socket directory.
// Priority: $ZMX_DIR > $XDG_RUNTIME_DIR/zmx > $TMPDIR/zmx-{uid} > parse from zmosh/zmx version output.
func resolveZmxDir() (string, error) {
	if d := os.Getenv("ZMX_DIR"); d != "" {
		return d, nil
	}
	if d := os.Getenv("XDG_RUNTIME_DIR"); d != "" {
		candidate := d + "/zmx"
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}
	}
	if tmp := os.Getenv("TMPDIR"); tmp != "" {
		candidate := fmt.Sprintf("%szmx-%d", tmp, os.Getuid())
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}
	}
	// Fallback: parse socket_dir from zmosh/zmx version output
	for _, bin := range []string{"zmosh", "zmx"} {
		out, err := exec.Command(bin, "version").Output()
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(out), "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "socket_dir") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					return fields[len(fields)-1], nil
				}
			}
		}
	}
	return "", fmt.Errorf("could not resolve zmx socket directory")
}

// fastListDir reads session names directly from socket files in the zmx directory.
func fastListDir(dir string) ([]Session, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	active := os.Getenv("ZMX_SESSION")
	var sessions []Session
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.Mode().Type()&os.ModeSocket == 0 {
			continue
		}
		sessions = append(sessions, Session{
			Name:      e.Name(),
			StartedIn: "~",
			Active:    e.Name() == active,
		})
	}
	return sessions, nil
}

// FastList returns sessions by reading socket files directly, bypassing zmosh list.
// Falls back to List() if the socket directory cannot be resolved.
func FastList() ([]Session, error) {
	dir, err := resolveZmxDir()
	if err != nil {
		return List()
	}
	sessions, err := fastListDir(dir)
	if err != nil {
		return List()
	}
	return sessions, nil
}

// Kill runs `zmosh kill <name>`.
// Output is suppressed; the picker displays its own status message.
func Kill(name string) error {
	return exec.Command("zmosh", "kill", name).Run()
}

// ExecCommand replaces the current process with the given command.
func ExecCommand(path string, argv []string) error {
	return syscall.Exec(path, argv, os.Environ())
}
