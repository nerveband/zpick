package main

import (
	"github.com/nerveband/zpick/internal/zmosh"
)

func runAttach(args []string) error {
	name := args[0]
	dir := ""

	// Parse --dir flag
	for i := 1; i < len(args); i++ {
		if args[i] == "--dir" && i+1 < len(args) {
			dir = args[i+1]
			break
		}
	}

	if dir != "" {
		return zmosh.AttachInDir(name, dir)
	}
	return zmosh.Attach(name)
}
