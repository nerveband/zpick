package main

import (
	"fmt"
	"os"

	"github.com/nerveband/zpick/internal/check"
)

func runCheck() error {
	// Check for --json flag
	jsonOutput := false
	for _, arg := range os.Args[2:] {
		if arg == "--json" {
			jsonOutput = true
		}
	}

	result := check.Run()

	if jsonOutput {
		j, err := result.JSON()
		if err != nil {
			return err
		}
		fmt.Println(j)
		return nil
	}

	// Default: guided output with install instructions for missing deps
	result.PrintGuide()
	return nil
}
