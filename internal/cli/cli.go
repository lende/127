package cli

import (
	"flag"
	"fmt"
	"io"
	"runtime"

	"github.com/lende/127/lib127"
)

const version = "0.3"

const (
	StatusSuccess = 0 // StatusSuccess is the status code returned on success.
	StatusFailure = 1 // StatusFailure is the status code returned on failure.
)

// App is a command-line interface to lib127.
type App struct {
	stdout, stderr io.Writer
}

// NewApp returns a new App with the given stdout and stderr writers for output.
func NewApp(stdout, stderr io.Writer) *App {
	return &App{stdout: stdout, stderr: stderr}
}

// Run runs the application with the given arguments. Returns 0 on success and 1
// on failure.
func (a *App) Run(args ...string) int {
	const usage = `127 is a tool for mapping hostnames to random loopback addresses.

Usage: 127 [option ...] [hostname]
Print IP mapped to hostname, assigning a random IP if no mapping exists.

Options:
`

	flags := flag.NewFlagSet("127", flag.ContinueOnError)
	flags.SetOutput(a.stderr)
	flags.Usage = func() {
		fmt.Fprint(flags.Output(), usage)
		flags.PrintDefaults()
	}

	var (
		printVersion  = flags.Bool("v", false, "print version information")
		deleteMapping = flags.Bool("d", false, "delete mapping")
		hostsFile     = flags.String("f", lib127.DefaultHostsFile(), "path to hosts file")
		addressBlock  = flags.String("b", lib127.DefaultAddressBlock, "address block")
	)

	if err := flags.Parse(args); err != nil {
		return 1
	}

	if *printVersion {
		fmt.Fprintf(a.stdout, "127 %s %s/%s\n", version, runtime.GOOS, runtime.GOARCH)
		return 0
	}

	hosts := new(lib127.Hosts).
		WithHostsFile(*hostsFile).
		WithAddressBlock(*addressBlock)

	switch hostname := flags.Arg(0); {
	case hostname == "":
		return a.output(hosts.RandomIP())
	case *deleteMapping:
		return a.output(hosts.Remove(hostname))
	default:
		return a.output(hosts.Set(hostname))
	}
}

func (a *App) output(ip string, err error) int {
	if err != nil {
		fmt.Fprintf(a.stderr, "Error: %v\n", err)

		return 1
	}

	if ip != "" {
		fmt.Fprintln(a.stdout, ip)
	}

	return 0
}
