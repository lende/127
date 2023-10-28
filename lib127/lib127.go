// Package lib127 provides methods for mapping hostnames to random loopback
// addresses.
package lib127

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math/big"
	"net"

	"github.com/lende/127/internal/hosts"
)

// Hosts provide methods for mapping hostnames to random IP addresses. Its zero
// value is usable and provides good defaults.
type Hosts struct {
	filename string
	randFunc func(uint32) (uint32, error)
}

func NewHosts(filename string) *Hosts {
	return &Hosts{filename: filename}
}

// DefaultHosts is the default Hosts and is used by GetIP, RandomIP, Remove, and
// Set.
var DefaultHosts = &Hosts{}

// RandomIP returns a random unassigned loopback address.
func (h *Hosts) RandomIP() (ip string, err error) {
	f, err := h.hostsFile()
	if err != nil {
		return "", err
	}
	return h.randomIP(f)
}

// RandomIP returns a random unassigned loopback address.
//
// RandomIP is a wrapper around DefaultHosts.RandomIP.
func RandomIP() (ip string, err error) {
	return DefaultHosts.RandomIP()
}

// Set maps the specified hostname to a random unnasigned loopback address, and
// returns that IP. If the hostname is already mapped, we return the already
// assigned IP address instead.
func (h *Hosts) Set(hostname string) (ip string, err error) {
	f, err := h.hostsFile()
	if err != nil {
		return "", err
	}
	if ip, err = f.GetIP(hostname); ip != "" || err != nil {
		return ip, err
	}
	if ip, err = h.randomIP(f); err != nil {
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

// Set maps the specified hostname to an random unassigned loopback address, and
// returns that IP. If the hostname is already mapped, we return the already
// assigned IP address instead.
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

func (h *Hosts) hostsFile() (*hosts.File, error) {
	if h.filename == "" {
		h.filename = hosts.FileLocation
	}
	return hosts.Open(h.filename)
}

func (h *Hosts) randUint32(max uint32) (uint32, error) {
	if h.randFunc == nil {
		return defaultRandFunc(max)
	}
	return h.randFunc(max)
}

func (h *Hosts) randomIP(f *hosts.File) (string, error) {
	ip := make(net.IP, 4)

	for {
		// Generate a random offset.
		offset, err := h.randUint32(maxIP - minIP)
		if err != nil {
			return "", fmt.Errorf("lib127: cound not generate random offset: %w", err)
		}

		// Add random offset and convert integer to IP address.
		binary.BigEndian.PutUint32(ip, minIP+offset)
		if f.HasIP(ip.String()) {
			continue
		}

		return ip.String(), nil
	}
}

// defaultRandFunc is a cryptographically secure random number generator.
func defaultRandFunc(max uint32) (uint32, error) {
	bigInt, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, err
	}
	return uint32(bigInt.Int64()), nil
}

const (
	minIP uint32 = 2130706434 // 127.0.0.2
	maxIP uint32 = 2147483646 // 127.255.255.254
)
