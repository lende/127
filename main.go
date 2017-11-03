// 127 is a simple tool for mapping host names to random loopback addresses.
package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/lende/127/lib"
)

const version = "0.2"

const usage = `127 is a tool for mapping hostnames to random loopback addresses.

Usage: 127 [option ...] [hostname[:port]] [operation]

Prints an unassigned random IP if hostname is left out.

Operations:

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
	printVersion, n :=
		flag.Bool("version", false, "print version information"),
		flag.Bool("n", false, "do not output a trailing newline")
	flag.Parse()

	if *printVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	var hostname, port, op string
	if hostname, op = flag.Arg(0), flag.Arg(1); hostname == "" {
		op = "ip"
	} else if h, p, err := net.SplitHostPort(hostname); err == nil {
		hostname, port = h, ":"+p
	}

	var ip string
	var err error
	switch op {
	case "ip":
		ip, err = lib127.RandomIP()
	case "set", "":
		ip, err = lib127.Set(hostname)
	case "get":
		ip, err = lib127.GetIP(hostname)
	case "remove":
		ip, err = lib127.Remove(hostname)
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown operation: %v\n", op)
		flag.Usage()
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(ip + port)
	if !*n && ip != "" {
		fmt.Print("\n")
	}
}
