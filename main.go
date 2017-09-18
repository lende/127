// 127 is a simple tool for mapping host names to random loopback addresses.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/lende/127/lib"
)

const version = "0.1.1"

const usage = `127 is a tool for mapping hostnames to random loopback addresses.

Usage: 127 [option ...] hostname [operation]

The operations are:

  set
        map hostname to random IP and print IP address (default)
  get
        print IP address associated with hostname
  remove
        remove hostname mapping

Options:

`

func main() {

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
		flag.PrintDefaults()
		fmt.Fprint(os.Stderr, "\n")
		os.Exit(1)
	}

	flag.StringVar(&lib127.HostsFile, "hosts", lib127.HostsFile, "path to hosts file")
	flag.StringVar(&lib127.AddressBlock, "block", lib127.AddressBlock, "address block")
	printVersion := flag.Bool("version", false, "print version information")
	n := flag.Bool("n", false, "do not output a trailing newline")

	flag.Parse()

	if *printVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	hostname, op := flag.Arg(0), flag.Arg(1)

	if hostname == "" {
		fmt.Fprint(os.Stderr, "missing hostname\n")
		flag.Usage()
	}

	var port string
	var ip string
	var err error

	if parts := strings.SplitN(hostname, ":", 2); len(parts) == 2 {
		hostname, port = parts[0], ":"+parts[1]
	}

	switch op {
	case "set", "":
		ip, err = lib127.Set(hostname)
	case "get":
		ip, err = lib127.Get(hostname)
	case "remove":
		ip, err = lib127.Remove(hostname)
	default:
		fmt.Fprintf(os.Stderr, "unknown operation: %v\n", op)
		flag.Usage()
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(ip + port)
	if !*n {
		fmt.Print("\n")
	}
}
