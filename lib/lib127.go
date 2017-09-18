// Package lib27 provides methods for mapping hostnames to random loopback addresses.
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

// Set maps the specified hostname to a random IP within the globally defind
// AddressBlock, and returns that address. If the hostname is already mapped, we
// return that IP address instead.
func Set(hostname string) (ip string, err error) {
	if hostname, err = adaptHostname(hostname); err != nil {
		return "", err
	}
	h, err := open(HostsFile)
	if err != nil {
		return "", err
	} else if ip, err = get(h, hostname); ip != "" || err != nil {
		return ip, err
	} else if ip, err = randomIP(h); err != nil {
		return "", err
	}
	return commit(h, ip, add(&h, hostname, ip))
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

// Remove removes the specified hostname and returns the associated IP. Returns
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
	return commit(h, ip, remove(&h, hostname, ip))
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

// These functions may be wrapped or replaced with build-specific code, or for
// testing and debugging purposes.
var (
	// open opens the HostsFile and returns a representation.
	open = func(path string) (h hostsfile.Hostsfile, err error) {
		f, err := os.Open(path)
		if err != nil {
			return h, err
		}
		defer f.Close()
		return hostsfile.Decode(f)
	}

	// add maps hostname to ip.
	add = func(h *hostsfile.Hostsfile, hostname, ip string) error {
		return h.Set(net.IPAddr{IP: net.ParseIP(ip)}, hostname)
	}

	// get returns the ip address associated with the hostname, if any.
	get = func(h hostsfile.Hostsfile, hostname string) (ip string, err error) {
		for _, r := range h.Records() {
			if r.Hostnames[hostname] {
				return r.IpAddress.String(), nil
			}
		}
		return "", nil
	}

	// remove removes hostname mapping.
	remove = func(h *hostsfile.Hostsfile, hostname, ip string) error {
		h.Remove(hostname)
		return nil
	}

	// fileWriter wraps the HostsFile for writing (used on Windows to replace "\n"
	// with "\r\n").
	fileWriter = func(f *os.File) io.Writer { return f }

	// adaptHostname validates hostname and converts from unicode to IDNA Punycode.
	adaptHostname = func(hostname string) (h string, err error) {
		h, err = idna.Lookup.ToASCII(hostname)
		if err != nil || net.ParseIP(hostname) != nil {
			return "", fmt.Errorf("invalid hostname: %v", hostname)
		}
		return h, nil
	}

	// ipSpan returns the smallest and largest valid IP (as integers) within the
	// specified ipnet.
	ipSpan = func(ipnet *net.IPNet) (minIP, maxIP uint32) {
		minIP = binary.BigEndian.Uint32(ipnet.IP.To4()) + 1
		maxIP = minIP + (^binary.BigEndian.Uint32(ipnet.Mask)) - 2
		return minIP, maxIP
	}

	// ips returns the set of all mapped IP addresses within the ipnet (used
	// to check for uniqueness).
	ips = func(h hostsfile.Hostsfile, ipnet *net.IPNet) map[string]bool {
		ips := make(map[string]bool)
		// Make sure we never reuse "localhost" (may be missing in hosts-file on Windows)
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

	// randomIP returns an unnasigned random IP within the AddressBlock.
	randomIP = func(h hostsfile.Hostsfile) (ip string, err error) {
		_, ipnet, err := net.ParseCIDR(AddressBlock)
		if err != nil {
			return "", err
		}
		if ones, _ := ipnet.Mask.Size(); ones > 30 {
			return "", fmt.Errorf("address block too small: %v", AddressBlock)
		}
		minIP, maxIP := ipSpan(ipnet)
		taken, netIP := ips(h, ipnet), make(net.IP, 4)
		if len(taken) >= int(maxIP-minIP) {
			return "", fmt.Errorf("no unnasigned IPs in address block: %v", AddressBlock)
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

	// commit commits changes to the HostsFile.
	commit = func(h hostsfile.Hostsfile, ip string, err error) (string, error) {
		if err != nil {
			return "", err
		}
		f, err := os.OpenFile(HostsFile, os.O_WRONLY|os.O_TRUNC, 0)
		if err != nil {
			return "", err
		}
		defer f.Close()
		if err = hostsfile.Encode(fileWriter(f), h); err != nil {
			return "", err
		}
		return ip, nil
	}
)
