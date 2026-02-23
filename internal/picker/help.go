package picker

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nerveband/zpick/internal/backend"
	"golang.org/x/term"
)

// showHelpConfig renders the help/config screen on tty.
// Handles 'b' to cycle backend, 'u' to toggle UDP. Esc returns to picker.
func showHelpConfig(tty *os.File, b backend.Backend, version string) {
	for {
		renderHelp(tty, b, version)

		oldState, err := term.MakeRaw(int(tty.Fd()))
		if err != nil {
			return
		}

		buf := make([]byte, 3)
		n, err := tty.Read(buf)
		term.Restore(int(tty.Fd()), oldState)

		if err != nil {
			return
		}

		key := buf[0]

		// Esc or Ctrl-C → back to picker
		if n == 1 && (key == 27 || key == 3) {
			return
		}

		switch key {
		case 'b':
			b = cycleBackend(tty, b)
		case 'u':
			toggleUDP(tty)
		}
	}
}

func renderHelp(tty *os.File, b backend.Backend, version string) {
	// Clear screen
	fmt.Fprint(tty, "\033[2J\033[H")

	fmt.Fprintf(tty, "  %szp help%s\n\n", boldCyan, reset)

	// Keys section
	fmt.Fprintf(tty, "  %sKeys%s\n", boldWht, reset)
	fmt.Fprintf(tty, "    %s1-9,a-y%s  attach session       %senter%s  new session\n", boldYel, reset, boldGrn, reset)
	fmt.Fprintf(tty, "    %sc%s        custom name           %sd%s      +date name\n", magenta, reset, cyan, reset)
	fmt.Fprintf(tty, "    %sz%s        pick dir (zoxide)     %sk%s      kill session\n", magenta, reset, red, reset)
	fmt.Fprintf(tty, "    %sh%s        this screen           %sesc%s    skip\n", cyan, reset, yellow, reset)
	fmt.Fprintln(tty)

	// Config section
	configDir := backend.ConfigDir()
	home, _ := os.UserHomeDir()
	displayDir := strings.Replace(configDir, home, "~", 1)

	fmt.Fprintf(tty, "  %sConfig%s %s%s/%s\n", boldWht, reset, dim, displayDir, reset)

	// Backend
	available := backend.Detect()
	availStr := strings.Join(available, ", ")
	fmt.Fprintf(tty, "    %sb%s  backend    %s%-12s%s %s[%s]%s\n",
		magenta, reset, boldWht, b.Name(), reset, dim, availStr, reset)

	// Guard
	apps := readGuardApps(configDir)
	appsStr := strings.Join(apps, ", ")
	fmt.Fprintf(tty, "    %s·%s  guard      %s%s%s\n", dim, reset, dim, appsStr, reset)
	fmt.Fprintf(tty, "    %s%s  manage:    zp guard add/remove/list%s\n", dim, dim, reset)

	// UDP
	udpEnabled, udpHost := backend.ReadUDP()
	udpStr := "off"
	if udpEnabled {
		udpStr = "on"
		if udpHost != "" {
			udpStr += " (" + udpHost + ")"
		} else {
			udpStr += " (no host)"
		}
	}
	if b.Name() == "zmosh" {
		fmt.Fprintf(tty, "    %su%s  udp        %s%s%s\n", magenta, reset, dim, udpStr, reset)
	} else {
		fmt.Fprintf(tty, "    %s·%s  udp        %s%s (zmosh only)%s\n", dim, reset, dim, udpStr, reset)
	}

	fmt.Fprintln(tty)
	fmt.Fprintf(tty, "  %s%s%s  %sgithub.com/nerveband/zpick%s\n", dim, version, reset, dim, reset)
	fmt.Fprintf(tty, "  %sesc%s %sback%s\n", yellow, reset, dim, reset)
}

func cycleBackend(tty *os.File, current backend.Backend) backend.Backend {
	available := backend.Detect()
	if len(available) < 2 {
		return current
	}

	// Find current index and cycle to next
	currentName := current.Name()
	nextIdx := 0
	for i, name := range available {
		if name == currentName {
			nextIdx = (i + 1) % len(available)
			break
		}
	}

	nextName := available[nextIdx]
	if err := backend.SetBackend(nextName); err != nil {
		fmt.Fprintf(tty, "\r  %sfailed: %v%s", dim, err, reset)
		return current
	}

	// Load the new backend
	newB, err := backend.Load(false)
	if err != nil {
		fmt.Fprintf(tty, "\r  %sfailed: %v%s", dim, err, reset)
		return current
	}

	return newB
}

func toggleUDP(tty *os.File) {
	enabled, host := backend.ReadUDP()
	if err := backend.SetUDP(!enabled, host); err != nil {
		fmt.Fprintf(tty, "\r  %sfailed: %v%s", dim, err, reset)
	}
}

// readGuardApps reads guard.conf from the config dir.
// Returns defaults if the file doesn't exist.
func readGuardApps(configDir string) []string {
	data, err := os.ReadFile(filepath.Join(configDir, "guard.conf"))
	if err != nil {
		return []string{"claude", "codex", "opencode"}
	}
	var apps []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			apps = append(apps, line)
		}
	}
	if len(apps) == 0 {
		return []string{"claude", "codex", "opencode"}
	}
	return apps
}
