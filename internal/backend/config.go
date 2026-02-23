package backend

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// validBackends is the list of recognized backend names.
var validBackends = []string{"zmosh", "zmx", "tmux", "shpool"}

// configDir returns the zpick config directory, respecting XDG_CONFIG_HOME.
func configDir() string {
	if d := os.Getenv("XDG_CONFIG_HOME"); d != "" {
		return filepath.Join(d, "zpick")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "zpick")
}

// ReadBackendName returns the configured backend name, or empty if not configured.
func ReadBackendName() (string, error) {
	return readBackendConfig()
}

// readBackendConfig reads the backend name from the config file.
// Returns empty string if the file doesn't exist.
func readBackendConfig() (string, error) {
	data, err := os.ReadFile(filepath.Join(configDir(), "backend"))
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// SetBackend writes the backend name to the config file.
func SetBackend(name string) error {
	if !isValidBackend(name) {
		return fmt.Errorf("unknown backend %q (valid: %s)", name, strings.Join(validBackends, ", "))
	}
	dir := configDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "backend"), []byte(name+"\n"), 0644)
}

// SetUDP writes the zmosh UDP configuration.
func SetUDP(enabled bool, host string) error {
	dir := configDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	var b strings.Builder
	if enabled {
		b.WriteString("enabled=true\n")
	} else {
		b.WriteString("enabled=false\n")
	}
	if host != "" {
		fmt.Fprintf(&b, "host=%s\n", host)
	}
	return os.WriteFile(filepath.Join(dir, "udp.conf"), []byte(b.String()), 0644)
}

// ReadUDP reads the zmosh UDP configuration.
// Defaults: enabled=true, host="" (empty).
func ReadUDP() (enabled bool, host string) {
	data, err := os.ReadFile(filepath.Join(configDir(), "udp.conf"))
	if err != nil {
		return true, "" // default: enabled, no host
	}
	enabled = true // default
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if k, v, ok := strings.Cut(line, "="); ok {
			switch k {
			case "enabled":
				enabled = v == "true"
			case "host":
				host = v
			}
		}
	}
	return
}

// Detect returns the names of all available backends (binaries found in PATH).
func Detect() []string {
	var found []string
	for _, name := range validBackends {
		if _, err := exec.LookPath(name); err == nil {
			found = append(found, name)
		}
	}
	if found == nil {
		found = []string{} // never return nil
	}
	return found
}

// Load reads the backend config and returns the appropriate Backend instance.
// When interactive=true and no config exists with multiple backends available,
// it prompts the user on /dev/tty.
// When interactive=false, it auto-detects a single backend or returns an error.
func Load(interactive bool) (Backend, error) {
	name, err := readBackendConfig()
	if err != nil {
		return nil, fmt.Errorf("reading backend config: %w", err)
	}

	if name != "" {
		return newBackend(name)
	}

	// No config — detect available backends
	available := Detect()

	switch len(available) {
	case 0:
		return nil, fmt.Errorf("no supported session manager found (install zmosh, zmx, tmux, or shpool)")
	case 1:
		// Auto-select the only available backend and save it
		if err := SetBackend(available[0]); err != nil {
			// Non-fatal — just use it without saving
		}
		return newBackend(available[0])
	default:
		if !interactive {
			return nil, fmt.Errorf("multiple backends available (%s); run 'zp check' to select one", strings.Join(available, ", "))
		}
		return promptBackend(available)
	}
}

// registry maps backend names to their factory functions.
// Populated by each backend's init() function.
var registry = map[string]func() Backend{}

// Register adds a backend factory to the registry.
// Called by each backend package's init() function.
func Register(name string, factory func() Backend) {
	registry[name] = factory
}

// newBackend creates a Backend by name from the registry.
func newBackend(name string) (Backend, error) {
	factory, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unknown backend %q (not registered)", name)
	}
	return factory(), nil
}

// promptBackend prompts the user to select a backend on /dev/tty.
func promptBackend(available []string) (Backend, error) {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("cannot open /dev/tty for backend selection: %w", err)
	}
	defer tty.Close()

	fmt.Fprintf(tty, "\n  Multiple session managers found. Pick one:\n\n")
	for i, name := range available {
		fmt.Fprintf(tty, "    %d) %s\n", i+1, name)
	}
	fmt.Fprintf(tty, "\n  > ")

	buf := make([]byte, 4)
	n, err := tty.Read(buf)
	if err != nil || n == 0 {
		return nil, fmt.Errorf("no selection made")
	}

	choice := int(buf[0] - '0')
	if choice < 1 || choice > len(available) {
		return nil, fmt.Errorf("invalid selection")
	}

	selected := available[choice-1]
	if err := SetBackend(selected); err != nil {
		// Non-fatal
	}
	fmt.Fprintf(tty, "  Using %s\n\n", selected)
	return newBackend(selected)
}

func isValidBackend(name string) bool {
	for _, v := range validBackends {
		if v == name {
			return true
		}
	}
	return false
}
