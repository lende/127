// Package lib27 provides methods for mapping hostnames to random loopback
// addresses.
package lib127

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/lende/127/internal/hostsfile"
)

// Default address block.
var AddressBlock = "127.0.0.0/8"

// OS-specific hosts-file location.
var HostsFile = hostsfile.Location

// RandomIP returns an unassigned random ip within the AddressBlock.
func RandomIP() (ip string, err error) {
	h, err := hostsfile.Open(HostsFile)
	if err != nil {
		return "", err
	}
	return randomIP(h, AddressBlock)
}

// Set maps the specified hostname to an unnasigned random IP within the
// AddressBlock, and returns that IP. If the hostname is already mapped, we
// return the already assigned IP address instead.
func Set(hostname string) (ip string, err error) {
	h, err := hostsfile.Open(HostsFile)
	if err != nil {
		return "", err
	}
	if ip, err = h.GetIP(hostname); ip != "" || err != nil {
		return ip, err
	}
	if ip, err = randomIP(h, AddressBlock); err != nil {
		return "", err
	}
	if err = h.Set(hostname, ip); err != nil {
		return "", err
	}
	return ip, h.Save()
}

// GetIP gets the IP associated with the specified hostname. Returns the empty
// string if hostname were not found.
func GetIP(hostname string) (ip string, err error) {
	h, err := hostsfile.Open(HostsFile)
	if err != nil {
		return "", err
	}
	return h.GetIP(hostname)
}

// Remove unmaps the specified hostname and returns the associated IP. Returns
// the empty string if hostname were not found.
func Remove(hostname string) (ip string, err error) {
	h, err := hostsfile.Open(HostsFile)
	if err != nil {
		return "", err
	}
	if ip, err = h.GetIP(hostname); err != nil {
		return "", err
	}
	if err = h.Remove(hostname); err != nil {
		return "", err
	}
	return ip, h.Save()
}

// ipSpan returns the smallest and largest valid IP (as integers) within the
// specified IP network.
func ipSpan(ipnet *net.IPNet) (minIP, maxIP uint32) {
	minIP = binary.BigEndian.Uint32(ipnet.IP.To4()) + 1
	maxIP = minIP + (^binary.BigEndian.Uint32(ipnet.Mask)) - 2
	return minIP, maxIP
}

// ips returns the set of all mapped IP addresses within the given IP network
// (used to check for uniqueness).
func ips(h *hostsfile.Hostsfile, ipnet *net.IPNet) map[string]bool {
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

func init() {
	rand.Seed(time.Now().UnixNano())
}

// randomIP returns an unnasigned random IP within the given address block.
func randomIP(h *hostsfile.Hostsfile, block string) (ip string, err error) {
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
