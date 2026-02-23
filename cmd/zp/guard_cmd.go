package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/nerveband/zpick/internal/guard"
	"github.com/nerveband/zpick/internal/hook"
)

func runGuard(args []string) error {
	// Handle management flags (don't need a backend)
	for i, arg := range args {
		switch arg {
		case "--add":
			if i+1 >= len(args) {
				return fmt.Errorf("--add requires an app name")
			}
			name := args[i+1]
			if err := guard.AddApp(name); err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "  added %q to guard list\n", name)
			if err := hook.Install(); err != nil {
				fmt.Fprintf(os.Stderr, "  warning: could not update hook: %v\n", err)
			}
			fmt.Fprintln(os.Stderr, "  restart your shell or run: source ~/.zshrc")
			return nil

		case "--remove":
			if i+1 >= len(args) {
				return fmt.Errorf("--remove requires an app name")
			}
			name := args[i+1]
			if err := guard.RemoveApp(name); err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "  removed %q from guard list\n", name)
			if err := hook.Install(); err != nil {
				fmt.Fprintf(os.Stderr, "  warning: could not update hook: %v\n", err)
			}
			fmt.Fprintln(os.Stderr, "  restart your shell or run: source ~/.zshrc")
			return nil

		case "--list":
			apps, err := guard.ReadConfig()
			if err != nil {
				return err
			}
			for _, app := range apps {
				fmt.Println(app)
			}
			return nil
		}
	}

	// Extract argv after "--"
	var argv []string
	for i, arg := range args {
		if arg == "--" {
			argv = args[i+1:]
			break
		}
	}

	// Load backend for guard prompt
	b, err := loadBackend(true)
	if err != nil {
		return err
	}

	cmd, err := guard.Run(b, argv)
	if err != nil {
		return err
	}
	if cmd != "" {
		fmt.Print(cmd)
	}
	return nil
}

func guardUsage() string {
	return strings.TrimSpace(`
Usage:
  zp guard -- <command> [args...]   Show session prompt before running command
  zp guard --add <app>              Add app to guard list
  zp guard --remove <app>           Remove app from guard list
  zp guard --list                   List guarded apps`)
}
