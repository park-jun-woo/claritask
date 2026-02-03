package main

import (
	"os"

	"parkjunwoo.com/claritask/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
