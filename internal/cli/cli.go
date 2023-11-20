package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/lende/127/lib127"
)

// Status codes returned by App to indicate sucess or failure.
const (
	StatusSuccess = 0
	StatusFailure = 1
)

// App is a command-line interface to lib127.
type App struct {
	Name, Version       string
	Writer, ErrorWriter io.Writer
}

// Run runs the application with the given arguments. Returns 0 on success and 1
// on failure.
func (a App) Run(args ...string) int {
	cmd := command{filename: lib127.DefaultHostsFile}
	if ok := a.parse(args, &cmd); !ok {
		return StatusFailure
	}

	return a.exec(cmd)
}

func (a App) name() string {
	if a.Name != "" {
		return a.Name
	}
	return filepath.Base(os.Args[0])
}

func (a App) version() string {
	if a.Version != "" {
		return a.Version
	}
	return "0.0.0-dev"
}

func (a App) writer() io.Writer {
	if a.Writer != nil {
		return a.Writer
	}
	return os.Stdout
}

func (a App) errorWriter() io.Writer {
	if a.ErrorWriter != nil {
		return a.ErrorWriter
	}
	return os.Stderr
}

type command struct {
	printVersion       bool
	filename, hostname string
	unmap, echo        bool
}

func (a App) parse(args []string, cmd *command) bool {
	const usageFmt = `%s is a tool for mapping hostnames to random loopback addresses.

Usage: %s [option ...] [hostname]
Print IP mapped to hostname, assigning a random IP if no mapping exists.

Options:
`

	flags := flag.NewFlagSet("127", flag.ContinueOnError)
	flags.SetOutput(a.errorWriter())
	flags.Usage = func() {
		fmt.Fprintf(a.errorWriter(), usageFmt, a.name(), a.name())
		flags.PrintDefaults()
	}

	flags.BoolVar(&cmd.printVersion, "v", false, "print version")
	flags.StringVar(&cmd.filename, "f", lib127.DefaultHostsFile, "path to hosts file")
	flags.BoolVar(&cmd.echo, "e", false, "echo hostname")
	flags.BoolVar(&cmd.unmap, "u", false, "unmap hostname")

	if err := flags.Parse(args); err != nil {
		return false
	}

	cmd.hostname = flags.Arg(0)
	return true
}

func (a App) exec(cmd command) int {
	if cmd.printVersion {
		fmt.Fprintf(a.writer(), "%s %s %s/%s\n",
			a.name(), a.version(), runtime.GOOS, runtime.GOARCH)
		return StatusSuccess
	}

	hosts, err := lib127.Open(cmd.filename)
	if err != nil {
		return a.error(cmd, err)
	}

	var host string
	switch {
	case cmd.hostname == "":
		host, err = hosts.RandomIP()
	case cmd.unmap:
		host, err = hosts.Unmap(cmd.hostname)
	default:
		host, err = hosts.Map(cmd.hostname)
	}

	if err != nil {
		return a.error(cmd, err)
	}

	if err := hosts.Save(); err != nil {
		return a.error(cmd, err)
	}

	if cmd.echo {
		host = cmd.hostname
	}
	fmt.Fprintln(a.writer(), host)

	return StatusSuccess
}

func (a App) error(cmd command, err error) int {
	var pathErr *fs.PathError
	switch {
	case errors.Is(err, lib127.ErrHostnameIsIP):
		// Echo IP addresses to stdout instead of failing with an error.
		fmt.Fprintln(a.writer(), cmd.hostname)
		return StatusSuccess
	case errors.Is(err, lib127.ErrHostnameInvalid):
		fmt.Fprintf(a.errorWriter(), "%s: invalid hostname: %s\n", a.name(), cmd.hostname)
	case errors.Is(err, lib127.ErrCannotUnmapLocalhost):
		fmt.Fprintf(a.errorWriter(), "%s: cannot remove localhost\n", a.name())
	case errors.As(err, &pathErr):
		fmt.Fprintf(a.errorWriter(), "%s: %v\n", a.name(), pathErr)
	default:
		fmt.Fprintf(a.errorWriter(), "%s: %s\n",
			a.name(), strings.TrimPrefix(err.Error(), "lib127: "))
	}

	return StatusFailure
}
