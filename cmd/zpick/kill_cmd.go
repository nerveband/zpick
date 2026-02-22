package main

import (
	"github.com/nerveband/zpick/internal/zmosh"
)

func runKill(name string) error {
	return zmosh.Kill(name)
}
