package hook

import (
	"os"
	"path/filepath"
)

func bashrcPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".bashrc")
}
