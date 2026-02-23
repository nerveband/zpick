package hook

import (
	"os"
	"path/filepath"
)

func zshrcPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".zshrc")
}
