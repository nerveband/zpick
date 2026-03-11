package main

import (
	"fmt"

	"github.com/nerveband/zpick/internal/switcher"
)

// runResume reads the switch-target file and outputs the shell command
// to attach to the target session. Called by the shell hook via
// eval "$(command zp resume)".
func runResume() error {
	b, err := loadBackend(false)
	if err != nil {
		return err
	}

	target, err := switcher.Read()
	if err != nil {
		// No target (missing file, stale, etc.) — silent, not an error.
		return nil
	}

	cmd := b.AttachCommand(target.Name, "")

	switch target.Action {
	case "attach", "new":
		if target.Dir != "" {
			fmt.Printf("cd %q && ZPICK_SESSION=%q %s", target.Dir, target.Name, cmd)
		} else {
			fmt.Printf("ZPICK_SESSION=%q %s", target.Name, cmd)
		}
	default:
		// Unknown action — silent, not an error.
		return nil
	}

	return nil
}
