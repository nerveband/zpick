package hook

import (
	"fmt"
	"os"
	"path/filepath"
)

func bashrcPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".bashrc")
}

func installBash() error {
	path := bashrcPath()

	// Idempotency: don't add twice
	if hasHook(path) {
		fmt.Printf("  hook already installed in %s\n", path)
		appendTermFix(path)
		return nil
	}

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
