package main

import (
	"os"

	"github.com/pastel-sketchbook/reins/internal/cli"
)

// version is injected via ldflags at build time.
var version = "dev"

func main() {
	cli.SetVersion(version)
	os.Exit(cli.Run(os.Args))
}
