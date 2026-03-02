package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/nerveband/zpick/internal/guard"
	"github.com/nerveband/zpick/internal/hook"
)

func runGuard(args []string) error {
	// No args and no "--": print explanation
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, guardExplanation())
		return nil
	}

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
			if hook.HasGuardInstalled() {
				if err := hook.Install(true); err != nil {
					fmt.Fprintf(os.Stderr, "  warning: could not update hook: %v\n", err)
				}
				fmt.Fprintln(os.Stderr, "  restart your shell or run: source ~/.zshrc")
			} else {
				fmt.Fprintln(os.Stderr, "  run 'zp install-guard' to activate guard wrappers")
			}
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
			if hook.HasGuardInstalled() {
				if err := hook.Install(true); err != nil {
					fmt.Fprintf(os.Stderr, "  warning: could not update hook: %v\n", err)
				}
				fmt.Fprintln(os.Stderr, "  restart your shell or run: source ~/.zshrc")
			}
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

	// Check for "--" before loading backend
	hasDash := false
	for _, arg := range args {
		if arg == "--" {
			hasDash = true
			break
		}
	}

	if !hasDash {
		// No management flags and no "--": print explanation
		fmt.Fprint(os.Stderr, guardExplanation())
		return nil
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

func guardExplanation() string {
	return `Session guard — what it does and how it works

  What: Intercepts specific commands (like claude, codex, aider) and prompts
  you to pick a session before running them. If you're already in a session,
  the command runs normally.

  Why: AI coding tools can run for a long time. If your SSH connection drops
  or your terminal closes, you lose the running process. The guard nudges you
  into a persistent session first, so your work survives disconnects.

  How: Shell function wrappers shadow the guarded commands. When you type
  "claude", the wrapper checks if you're in a session. If not, it shows a
  quick prompt (10s timeout). Press Enter to pick a session, Esc to skip.

  Limitations:
    - Only works in interactive shells (the wrapper must be sourced)
    - Shadows the binary with a shell function (won't apply in scripts)
    - 10-second timeout before auto-skipping the prompt

  Commands:
    zp install-guard          Install guard wrappers into your shell config
    zp remove-guard           Remove guard wrappers (keeps the shell hook)
    zp guard --add <app>      Add an app to the guard list
    zp guard --remove <app>   Remove an app from the guard list
    zp guard --list           Show guarded apps

`
}

func guardUsage() string {
	return strings.TrimSpace(`
Usage:
  zp guard                            Explain what guard does
  zp guard -- <command> [args...]      Show session prompt before running command
  zp guard --add <app>                 Add app to guard list
  zp guard --remove <app>              Remove app from guard list
  zp guard --list                      List guarded apps`)
}
