package main

import (
	"fmt"
	"os"

	"github.com/nerveband/zpick/internal/backend"
	"github.com/nerveband/zpick/internal/update"

	// Register all backends via init()
	_ "github.com/nerveband/zpick/internal/backend/shpool"
	_ "github.com/nerveband/zpick/internal/backend/tmux"
	_ "github.com/nerveband/zpick/internal/backend/zmosh"
	_ "github.com/nerveband/zpick/internal/backend/zmx"
)

var version = "dev"

func main() {
	if len(os.Args) < 2 {
		if err := runPicker(); err != nil {
			fmt.Fprintf(os.Stderr, "zp: %v\n", err)
			os.Exit(1)
		}
		return
	}

	var updateCh <-chan update.CheckResult
	if shouldCheckUpdates(os.Args[1:]) {
		updateCh = update.CheckAsync(version)
	}

	switch os.Args[1] {
	case "list":
		if err := runList(); err != nil {
			fmt.Fprintf(os.Stderr, "zp: %v\n", err)
			os.Exit(1)
		}
	case "check":
		if err := runCheck(); err != nil {
			fmt.Fprintf(os.Stderr, "zp: %v\n", err)
			os.Exit(1)
		}
	case "attach":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: zp attach <name> [--dir <path>]")
			os.Exit(1)
		}
		if err := runAttach(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "zp: %v\n", err)
			os.Exit(1)
		}
	case "kill":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: zp kill <name>")
			os.Exit(1)
		}
		if err := runKill(os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "zp: %v\n", err)
			os.Exit(1)
		}
	case "guard":
		if err := runGuard(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "zp: %v\n", err)
			os.Exit(1)
		}
	case "autorun":
		if err := runAutorun(); err != nil {
			fmt.Fprintf(os.Stderr, "zp: %v\n", err)
			os.Exit(1)
		}
	case "install-hook":
		if err := runInstallHook(); err != nil {
			fmt.Fprintf(os.Stderr, "zp: %v\n", err)
			os.Exit(1)
		}
	case "upgrade":
		if err := runUpgrade(); err != nil {
			fmt.Fprintf(os.Stderr, "zp: %v\n", err)
			os.Exit(1)
		}
	case "version":
		fmt.Printf("zp %s\n", version)
	case "--help", "-h", "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "zp: unknown command %q\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}

	if updateCh != nil {
		select {
		case result := <-updateCh:
			if notice := update.FormatNotice(result); notice != "" {
				fmt.Fprint(os.Stderr, notice)
			}
		default:
		}
	}
}

func loadBackend(interactive bool) (backend.Backend, error) {
	return backend.Load(interactive)
}

func hasJSONFlag() bool {
	for _, arg := range os.Args[2:] {
		if arg == "--json" {
			return true
		}
	}
	return false
}

func shouldCheckUpdates(args []string) bool {
	if len(args) == 0 {
		return false
	}
	switch args[0] {
	case "version", "upgrade", "--help", "-h", "help", "guard", "autorun":
		return false
	}
	for _, arg := range args[1:] {
		if arg == "--json" {
			return false
		}
	}
	return true
}

func printUsage() {
	fmt.Println(`zp — session launcher

Usage:
  zp              Interactive TUI picker (default)
  zp list         List sessions (--json for machine-readable)
  zp check        Check dependencies (--json for machine-readable)
  zp attach <n>   Attach or create session
  zp kill <name>  Kill a session
  zp guard        Session guard for AI coding tools
  zp install-hook Add shell hook to .zshrc/.bashrc/.config/fish
  zp upgrade      Upgrade to the latest version
  zp version      Print version`)
}
