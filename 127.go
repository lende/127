// 127 is a simple tool for mapping host names to random loopback addresses.
package main

import (
	_ "embed"
	"os"
	"strings"

	"github.com/lende/127/internal/cli"
)

//go:embed VERSION
var version string

func main() {
	app := cli.App{Version: strings.TrimSpace(version)}
	os.Exit(app.Run(os.Args[1:]...))
}
