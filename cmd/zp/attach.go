package main

import (
	"os"
)

func runAttach(args []string) error {
	b, err := loadBackend(true)
	if err != nil {
		return err
	}

	name := args[0]
	dir := ""

	for i := 1; i < len(args); i++ {
		if args[i] == "--dir" && i+1 < len(args) {
			dir = args[i+1]
			break
		}
	}

	if dir != "" {
		if err := os.Chdir(dir); err != nil {
			return err
		}
	}
	return b.Attach(name)
}
