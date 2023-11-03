package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"runtime"
	"strings"

	"github.com/lende/127/lib127"
)

const (
	StatusSuccess = 0 // StatusSuccess is the status code returned on success.
	StatusFailure = 1 // StatusFailure is the status code returned on failure.
)

// App is a command-line interface to lib127.
type App struct {
	version        string
	stdout, stderr io.Writer
}

// NewApp returns a new App with the given stdout and stderr writers for output.
func NewApp(version string, stdout, stderr io.Writer) *App {
	return &App{version: version, stdout: stdout, stderr: stderr}
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
		echoHostname  = flags.Bool("e", false, "echo hostname")
		hostsFile     = flags.String("f", lib127.DefaultHostsFile, "path to hosts file")
	)

	if err := flags.Parse(args); err != nil {
		return 1
	}

	if *printVersion {
		fmt.Fprintf(a.stdout, "127 %s %s/%s\n", a.version, runtime.GOOS, runtime.GOARCH)
		return 0
	}

	hosts := lib127.NewHosts(*hostsFile)

	switch hostname := flags.Arg(0); {
	case hostname == "":
		return a.output(hosts.RandomIP())
	case *deleteMapping:
		return a.output(hosts.Remove(hostname))
	default:
		host, err := hosts.Set(hostname)
		if *echoHostname {
			host = hostname
		}
		return a.output(host, err)
	}
}

func (a *App) output(ip string, err error) int {
	if err != nil {
		return a.error(err)
	}

	if ip != "" {
		fmt.Fprintln(a.stdout, ip)
	}
	return 0
}

func (a *App) error(err error) int {
	var (
		pathErr *fs.PathError
		hostErr lib127.HostnameError
	)

	switch {
	case errors.Is(err, lib127.ErrCannotRemoveLocalhost):
		fmt.Fprintln(a.stderr, "127: cannot remove localhost")
	case errors.As(err, &pathErr):
		fmt.Fprintln(a.stderr, "127:", pathErr.Error())
	case errors.As(err, &hostErr):
		if errors.Is(err, lib127.ErrHostnameIsIP) {
			// Echo IP addresses to stdout instead of failing with an error.
			fmt.Fprintln(a.stdout, hostErr.Hostname())
			return 0
		}

		fmt.Fprintf(a.stderr, "127: invalid hostname: %s\n", hostErr.Hostname())
	default:
		fmt.Fprintln(a.stderr, "127: unexpected error:", strings.TrimPrefix(err.Error(), "lib127: "))
	}

	return 1
}
