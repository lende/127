// Package lib27 provides methods for mapping hostnames to random loopback
// addresses.
package lib127

import (
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"time"

	"github.com/kevinburke/hostsfile/lib"
	"golang.org/x/net/idna"
)

// Default address block.
var AddressBlock = "127.0.0.0/8"

// RandomIP returns an unassigned random ip within the AddressBlock.
func RandomIP() (ip string, err error) {
	h, err := open(HostsFile)
	if err != nil {
		return "", err
	}
	return randomIP(h, AddressBlock)
}

// Set maps the specified hostname to an unnasigned random IP within the
// AddressBlock, and returns that IP. If the hostname is already mapped, we
// return the already assigned IP address instead.
func Set(hostname string) (ip string, err error) {
	if hostname, err = adaptHostname(hostname); err != nil {
		return "", err
	}
	h, err := open(HostsFile)
	if err != nil {
		return "", err
	} else if ip, err = get(h, hostname); ip != "" || err != nil {
		return ip, err
	} else if ip, err = randomIP(h, AddressBlock); err != nil {
		return "", err
	}
	if err := add(&h, hostname, ip); err != nil {
		return "", err
	}
	return ip, commit(h, HostsFile)
}

// Get gets the IP associated with the specified hostname. Returns the empty
// string if hostname were not found.
func Get(hostname string) (ip string, err error) {
	if hostname, err = adaptHostname(hostname); err != nil {
		return "", err
	}
	h, err := open(HostsFile)
	if err != nil {
		return "", err
	}
	return get(h, hostname)
}

// Remove unmaps the specified hostname and returns the associated IP. Returns
// the empty string if hostname were not found.
func Remove(hostname string) (ip string, err error) {
	if hostname, err = adaptHostname(hostname); err != nil {
		return "", err
	}
	h, err := open(HostsFile)
	if err != nil {
		return "", err
	}
	ip, _ = get(h, hostname)
	if err := remove(&h, hostname, ip); err != nil {
		return "", err
	}
	return ip, commit(h, HostsFile)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

// These functions are defined as variables so they can be wrapped or redefined.
// Used for platform-specific code and for testing purposes. Some, in their
// default form, may be intentionally redundant or take unnused arguments.
var (
	// open opens the hosts-file and returns a representation.
	open = func(filename string) (h hostsfile.Hostsfile, err error) {
		f, err := os.Open(filename)
		if err != nil {
			return h, err
		}
		defer f.Close()
		return hostsfile.Decode(f)
	}

	// add maps the specified hostname to the given IP.
	add = func(h *hostsfile.Hostsfile, hostname, ip string) error {
		return h.Set(net.IPAddr{IP: net.ParseIP(ip)}, hostname)
	}

	// get returns the IP address associated with the given hostname, if any.
	get = func(h hostsfile.Hostsfile, hostname string) (ip string, err error) {
		for _, r := range h.Records() {
			if r.Hostnames[hostname] {
				return r.IpAddress.String(), nil
			}
		}
		return "", nil
	}

	// remove removes the given hostname mapping.
	remove = func(h *hostsfile.Hostsfile, hostname, ip string) error {
		h.Remove(hostname)
		return nil
	}

	// fileWriter wraps the hosts-file for writing (redefined on windows to produce
	// correct newline characters).
	fileWriter = func(f *os.File) io.Writer { return f }

	// adaptHostname validates the given hostname and converts it from unicode to
	// IDNA Punycode.
	adaptHostname = func(hostname string) (h string, err error) {
		h, err = idna.Lookup.ToASCII(hostname)
		if err != nil || net.ParseIP(hostname) != nil {
			return "", fmt.Errorf("invalid hostname: %v", hostname)
		}
		return h, nil
	}

	// ipSpan returns the smallest and largest valid IP (as integers) within the
	// specified IP network.
	ipSpan = func(ipnet *net.IPNet) (minIP, maxIP uint32) {
		minIP = binary.BigEndian.Uint32(ipnet.IP.To4()) + 1
		maxIP = minIP + (^binary.BigEndian.Uint32(ipnet.Mask)) - 2
		return minIP, maxIP
	}

	// ips returns the set of all mapped IP addresses within the given IP network
	// (used to check for uniqueness).
	ips = func(h hostsfile.Hostsfile, ipnet *net.IPNet) map[string]bool {
		ips := make(map[string]bool)
		// Make sure we never touch localhost (may be missing in hosts-file).
		if ipnet.Contains(net.IP{127, 0, 0, 1}) {
			ips["127.0.0.1"] = true
		}
		for _, r := range h.Records() {
			if ipnet.Contains(r.IpAddress.IP) {
				ips[r.IpAddress.String()] = true
			}
		}
		return ips
	}

	// randomIP returns an unnasigned random IP within the given address block.
	randomIP = func(h hostsfile.Hostsfile, block string) (ip string, err error) {
		_, ipnet, err := net.ParseCIDR(block)
		if err != nil {
			return "", err
		}
		if ones, _ := ipnet.Mask.Size(); ones > 30 {
			return "", fmt.Errorf("address block too small: %v", block)
		}
		minIP, maxIP := ipSpan(ipnet)
		taken, netIP := ips(h, ipnet), make(net.IP, 4)
		if len(taken) >= int(maxIP-minIP) {
			return "", fmt.Errorf("no unnasigned IPs in address block: %v", block)
		}
		for {
			// Generate a random offset.
			offset := uint32(rand.Int63n(int64(maxIP - minIP)))
			// Add random offset and convert integer to IP address.
			binary.BigEndian.PutUint32(netIP, minIP+offset)
			if ip = netIP.String(); !taken[ip] {
				break
			}
		}
		return ip, nil
	}

	// commit commits changes to the given hosts-file.
	commit = func(h hostsfile.Hostsfile, filename string) error {
		f, err := os.OpenFile(filename, os.O_WRONLY|os.O_TRUNC, 0)
		if err != nil {
			return err
		}
		defer f.Close()
		if err = hostsfile.Encode(fileWriter(f), h); err != nil {
			return err
		}
		return nil
	}
)
