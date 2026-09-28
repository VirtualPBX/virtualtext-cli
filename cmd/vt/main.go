package main

import (
	"os"

	"github.com/VirtualPBX/virtualtext-cli/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
