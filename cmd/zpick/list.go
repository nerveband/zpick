package main

import (
	"fmt"
	"os"

	"github.com/nerveband/zpick/internal/zmosh"
)

func runList() error {
	// Check for --json flag
	jsonOutput := false
	for _, arg := range os.Args[2:] {
		if arg == "--json" {
			jsonOutput = true
		}
	}

	if jsonOutput {
		j, err := zmosh.ListJSON()
		if err != nil {
			return err
		}
		fmt.Println(j)
		return nil
	}

	// Human-readable output
	sessions, err := zmosh.List()
	if err != nil {
		return err
	}

	if len(sessions) == 0 {
		fmt.Println("  no sessions")
		return nil
	}

	for _, s := range sessions {
		status := "."
		if s.Active {
			status = "*"
		}
		fmt.Printf("  %s%s  (%d clients)  %s\n", status, s.Name, s.Clients, s.StartedIn)
	}
	return nil
}
