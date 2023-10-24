// Package lib27 provides methods for mapping hostnames to random loopback
// addresses.
package lib127

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math/big"
	"net"

	"github.com/lende/127/internal/hostsfile"
)

// DefaultAddressBlock is the default address block.
const DefaultAddressBlock = "127.0.0.0/8"

// DefaultHostsFile returns the location of the default hosts file.
func DefaultHostsFile() string {
	return hostsfile.Location
}

// Hosts provide methods for mapping hostnames to random IP addresses. Its zero
// value is usable and provides good defaults.
type Hosts struct {
	filename string
	block    string
	randFunc func(uint32) (uint32, error)
}

// DefaultHosts is the default Hosts and is used by GetIP, RandomIP, Remove, and
// Set.
var DefaultHosts = &Hosts{}

// WithHostsFile configures and returns a copy of Hosts with the given hosts file.
func (h Hosts) WithHostsFile(filename string) *Hosts {
	h.filename = filename
	return &h
}

// WithFile configures and returns a copy of Hosts with the given address block.
func (h Hosts) WithAddressBlock(block string) *Hosts {
	h.block = block
	return &h
}

// RandomIP returns an unassigned random ip within the configured address block.
func (h *Hosts) RandomIP() (ip string, err error) {
	f, err := h.hostsFile()
	if err != nil {
		return "", err
	}
	return h.randomIP(f, h.addressBlock())
}

// RandomIP returns an unassigned random ip within the configured address block.
//
// RandomIP is a wrapper around DefaultHosts.RandomIP.
func RandomIP() (ip string, err error) {
	return DefaultHosts.RandomIP()
}

// Set maps the specified hostname to an unnasigned random IP within the
// configured address block, and returns that IP. If the hostname is already
// mapped, we return the already assigned IP address instead.
func (h *Hosts) Set(hostname string) (ip string, err error) {
	f, err := h.hostsFile()
	if err != nil {
		return "", err
	}
	if ip, err = f.GetIP(hostname); ip != "" || err != nil {
		return ip, err
	}
	if ip, err = h.randomIP(f, h.addressBlock()); err != nil {
		return "", err
	}
	if err = f.Set(hostname, ip); err != nil {
		return "", err
	}
	if err := f.Save(); err != nil {
		return "", err
	}
	return ip, nil
}

// Set maps the specified hostname to an unnasigned random IP within the
// AddressBlock, and returns that IP. If the hostname is already mapped, we
// return the already assigned IP address instead.
//
// Set is a wrapper around DefaultHosts.Set.
func Set(hostname string) (ip string, err error) {
	return DefaultHosts.Set(hostname)
}

// GetIP gets the IP associated with the specified hostname. Returns the empty
// string if hostname were not found.
func (h *Hosts) GetIP(hostname string) (ip string, err error) {
	f, err := h.hostsFile()
	if err != nil {
		return "", err
	}
	return f.GetIP(hostname)
}

// GetIP gets the IP associated with the specified hostname. Returns the empty
// string if hostname were not found.
//
// GetIP is a wrapper around DefaultHosts.GetIP.
func GetIP(hostname string) (ip string, err error) {
	return DefaultHosts.GetIP(hostname)
}

// Remove unmaps the specified hostname and returns the associated IP. Returns
// the empty string if hostname were not found.
func (h *Hosts) Remove(hostname string) (ip string, err error) {
	f, err := h.hostsFile()
	if err != nil {
		return "", err
	}
	if ip, err = f.GetIP(hostname); ip == "" || err != nil {
		return "", err
	}
	if err = f.Remove(hostname); err != nil {
		return "", err
	}
	if err = f.Save(); err != nil {
		return "", err
	}
	return ip, nil
}

// Remove unmaps the specified hostname and returns the associated IP. Returns
// the empty string if hostname were not found.
//
// Remove is a wrapper around DefaultHosts.Remove.
func Remove(hostname string) (ip string, err error) {
	return DefaultHosts.Remove(hostname)
}

func (h *Hosts) hostsFile() (*hostsfile.Hostsfile, error) {
	if h.filename == "" {
		h.filename = hostsfile.Location
	}
	return hostsfile.Open(h.filename)
}

func (h *Hosts) addressBlock() string {
	if h.block == "" {
		return DefaultAddressBlock
	}
	return h.block
}

func (h *Hosts) randUint32(max uint32) (uint32, error) {
	if h.randFunc == nil {
		return defaultRandFunc(max)
	}
	return h.randFunc(max)
}

func (h *Hosts) randomIP(f *hostsfile.Hostsfile, block string) (ip string, err error) {
	_, ipnet, err := net.ParseCIDR(block)
	if err != nil {
		return "", fmt.Errorf("lib127: could not parse address block: %w", err)
	}
	if ones, _ := ipnet.Mask.Size(); ones > 30 {
		return "", fmt.Errorf("lib127: address block too small: %v", block)
	}
	minIP, maxIP := ipSpan(ipnet)
	taken, netIP := ips(f, ipnet), make(net.IP, 4)
	if len(taken) >= int(maxIP-minIP) {
		return "", fmt.Errorf("lib127: no unnasigned IPs in address block: %v", block)
	}
	for {
		// Generate a random offset.
		offset, err := h.randUint32(maxIP - minIP) // #nosec G404
		if err != nil {
			return "", fmt.Errorf("lib127: cound not generate random offset: %v", err)
		}

		// Add random offset and convert integer to IP address.
		binary.BigEndian.PutUint32(netIP, minIP+offset)
		if ip = netIP.String(); !taken[ip] {
			break
		}
	}
	return ip, nil
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
func ips(f *hostsfile.Hostsfile, ipnet *net.IPNet) map[string]bool {
	ips := make(map[string]bool)
	// Make sure we never touch localhost (may be missing in hosts-file).
	if ipnet.Contains(net.IP{127, 0, 0, 1}) {
		ips["127.0.0.1"] = true
	}
	for _, r := range f.Records() {
		if ipnet.Contains(r.IpAddress.IP) {
			ips[r.IpAddress.String()] = true
		}
	}
	return ips
}

// defaultRandFunc is a cryptographically secure random number generator.
var defaultRandFunc = func(max uint32) (uint32, error) {
	bigInt, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, err
	}
	return uint32(bigInt.Int64()), nil
}
