package hook

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const p10kInstantPromptMarker = "Powerlevel10k instant prompt"

func zshrcPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".zshrc")
}

func installZsh() error {
	path := zshrcPath()

	// Idempotency: don't add twice
	if hasHook(path) {
		fmt.Printf("  hook already installed in %s\n", path)
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("cannot read %s: %w", path, err)
	}

	content := string(data)

	// Handle Powerlevel10k: hook must go BEFORE instant prompt block
	if strings.Contains(content, p10kInstantPromptMarker) {
		return insertBeforeP10k(path, content)
	}

	// Normal case: append to end of file
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
	return nil
}

func insertBeforeP10k(path, content string) error {
	lines := strings.Split(content, "\n")
	var result []string
	inserted := false

	for _, line := range lines {
		if !inserted && strings.Contains(line, p10kInstantPromptMarker) {
			// Insert hook before p10k block
			result = append(result, hookLine)
			result = append(result, "")
			inserted = true
		}
		result = append(result, line)
	}

	if !inserted {
		// Fallback: append
		result = append(result, "", hookLine)
	}

	if err := os.WriteFile(path, []byte(strings.Join(result, "\n")), 0644); err != nil {
		return fmt.Errorf("cannot write %s: %w", path, err)
	}

	fmt.Printf("  installed hook in %s (before Powerlevel10k instant prompt)\n", path)
	return nil
}
