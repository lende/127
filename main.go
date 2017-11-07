// 127 is a simple tool for mapping host names to random loopback addresses.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/lende/127/lib"
)

const version = "0.2"

const usage = `127 is a tool for mapping hostnames to random loopback addresses.

Usage: 127 [option ...] [hostname]
Print IP mapped to hostname, assigning a random IP if no mapping exists.

Options:
`

func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
		flag.PrintDefaults()
		fmt.Fprint(os.Stderr, "\n")
	}

	flag.StringVar(&lib127.HostsFile, "hosts", lib127.HostsFile, "path to hosts file")
	flag.StringVar(&lib127.AddressBlock, "block", lib127.AddressBlock, "address block")
	printVersion, n, delete :=
		flag.Bool("v", false, "print version information"),
		flag.Bool("n", false, "do not output a trailing newline"),
		flag.Bool("d", false, "delete hostname mapping")
	flag.Parse()

	if *printVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	var ip string
	var err error
	if hostname := flag.Arg(0); hostname == "" {
		ip, err = lib127.RandomIP()
	} else if *delete {
		ip, err = lib127.Remove(hostname)
	} else {
		ip, err = lib127.Set(hostname)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(ip)
	if !*n && ip != "" {
		fmt.Print("\n")
	}
}
