package hook

import (
	"fmt"
	"os"
	"path/filepath"
)

func zshrcPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".zshrc")
}

func installZsh() error {
	path := zshrcPath()

	// Idempotency: don't add twice
	if hasHook(path) {
		fmt.Printf("  hook already installed in %s\n", path)
		appendTermFix(path)
		return nil
	}

	// Append to end of file so the hook runs after PATH is configured.
	// zpick renders its TUI to /dev/tty, so Powerlevel10k instant prompt
	// is not a concern — no special insertion order needed.
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("cannot open %s: %w", path, err)
	}
	defer f.Close()

	_, err = fmt.Fprintf(f, "\n%s\n", hookLine)
	if err != nil {
		return err
	}

	fmt.Printf("  installed hook in %s\n", path)
	appendTermFix(path)
	return nil
}
