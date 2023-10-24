// 127 is a simple tool for mapping host names to random loopback addresses.
package main

import (
	"os"

	"github.com/lende/127/internal/cli"
)

func main() {
	status := cli.NewApp(os.Stdout, os.Stderr).Run(os.Args[1:]...)
	os.Exit(status)
}
