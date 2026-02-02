package main

import (
	"os"

	"parkjunwoo.com/talos/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
