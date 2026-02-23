package zmosh

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/nerveband/zpick/internal/backend"
)

// ResolveZmxDir finds the zmx socket directory.
// Priority: $ZMX_DIR > $XDG_RUNTIME_DIR/zmx > $TMPDIR/zmx-{uid} > parse from zmosh/zmx version output.
func ResolveZmxDir() (string, error) {
	if d := os.Getenv("ZMX_DIR"); d != "" {
		return d, nil
	}
	if d := os.Getenv("XDG_RUNTIME_DIR"); d != "" {
		candidate := filepath.Join(d, "zmx")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}
	}
	if tmp := os.Getenv("TMPDIR"); tmp != "" {
		// [CR] Fix Linux TMPDIR path join bug — use filepath.Join instead of string concat
		candidate := filepath.Join(tmp, fmt.Sprintf("zmx-%d", os.Getuid()))
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

// FastListDir reads session names directly from socket files in the zmx directory.
func FastListDir(dir string) ([]backend.Session, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	active := os.Getenv("ZMX_SESSION")
	var sessions []backend.Session
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
		sessions = append(sessions, backend.Session{
			Name:      e.Name(),
			StartedIn: "~",
			Active:    e.Name() == active,
		})
	}
	return sessions, nil
}

// itoa is a helper for strconv.Itoa.
func itoa(i int) string {
	return strconv.Itoa(i)
}
