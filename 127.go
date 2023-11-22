// 127 is a simple tool for mapping host names to random loopback addresses.
package main

import (
	_ "embed"
	"os"

	"github.com/lende/127/internal/cli"
	ver "github.com/lende/127/internal/version"
)

var version string

func main() {
	app := cli.App{Version: ver.Semantic(version)}
	os.Exit(app.Run(os.Args[1:]...))
}
