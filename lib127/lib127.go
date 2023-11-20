// Package lib127 provides methods for mapping hostnames to random loopback
// addresses.
package lib127

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io/fs"
	"math/big"
	"net"

	"github.com/lende/127/lib127/internal/hosts"
)

// DefaultHostsFile is the default hosts file location.
const DefaultHostsFile = "/etc/hosts"

// These errors can be tested against using errors.Is. They are never returned
// directly.
var (
	// ErrHostnameInvalid indicates an invalid hostname.
	ErrHostnameInvalid = hosts.ErrHostnameInvalid

	// ErrHostnameIsIP indicates that an IP address was given as hostname.
	ErrHostnameIsIP = hosts.ErrHostnameIsIP

	// ErrCannotUnmapLocalhost indicates a request to unmap localhost.
	ErrCannotUnmapLocalhost = errors.New("127: cannot unmap localhost")
)

// Hosts provide methods for mapping hostnames to random IP addresses. Its zero
// value is usable and operates on the default hosts file.
type Hosts struct {
	file     *hosts.File
	changed  bool
	randFunc func(uint32) (uint32, error)
}

// NewHosts opens a new Hosts using the given file. If filename is "" the
// default hosts file is opened.
//
// Returned file system errors wrap *fs.PathError.
func Open(filename string) (*Hosts, error) {
	f, err := hosts.Open(filename)
	if err != nil {
		return nil, wrapError("open file", err)
	}
	return &Hosts{file: f}, nil
}

const (
	minIP uint32 = 2130706434 // 127.0.0.2
	maxIP uint32 = 2147483646 // 127.255.255.254
)

// RandomIP returns a random unassigned loopback address.
func (h *Hosts) RandomIP() (string, error) {
	ip := make(net.IP, net.IPv4len)

	for {
		// Generate a random offset.
		offset, err := h.randUint32(maxIP - minIP)
		if err != nil {
			return "", wrapError("generate random offset", err)
		}

		// Add random offset and convert integer to IP address.
		binary.BigEndian.PutUint32(ip, minIP+offset)
		if h.file.HasIP(ip.String()) {
			continue
		}

		return ip.String(), nil
	}
}

// IP returns the IP address associated with the specified hostname. Returns an
// empty string if hostname were not found.
//
// Returned hostname errors can be matched against ErrInvalidHostname and
// ErrHostnameIsIP.
func (h *Hosts) IP(hostname string) (string, error) {
	if isLocalhost(hostname) {
		return "127.0.0.1", nil
	}

	ip, err := h.file.IP(hostname)
	if err != nil {
		return "", wrapError("get IP", err)
	}
	return ip, nil
}

// Map maps the specified hostname to a random unnasigned loopback address, and
// returns that IP. If the hostname is already mapped, we return the already
// assigned IP address instead.
//
// Returned hostname errors can be matched against ErrInvalidHostname and
// ErrHostnameIsIP.
func (h *Hosts) Map(hostname string) (string, error) {
	if isLocalhost(hostname) {
		return "127.0.0.1", nil
	}

	if ip, err := h.IP(hostname); ip != "" || err != nil {
		return ip, err
	}

	ip, err := h.RandomIP()
	if err != nil {
		return "", err
	}

	if err = h.file.Map(hostname, ip); err != nil {
		return "", wrapError("set hostname", err)
	}
	h.changed = true

	return ip, nil
}

// Unmap unmaps the specified hostname and returns the associated IP. Returns
// an empty string if hostname were not found.
//
// Returned hostname errors can be matched against ErrInvalidHostname and
// ErrHostnameIsIP.
func (h *Hosts) Unmap(hostname string) (string, error) {
	if isLocalhost(hostname) {
		return "", ErrCannotUnmapLocalhost
	}

	ip, err := h.IP(hostname)
	if err != nil {
		return "", err
	}

	if err = h.file.Unmap(hostname); err != nil {
		return "", wrapError("set hostname", err)
	}
	h.changed = true

	return ip, nil
}

// Save saves the modified hosts file to disk. Does nothing if no changes were
// made.
//
// Returned file system errors wrap *fs.PathError.
func (h *Hosts) Save() error {
	if !h.changed {
		return nil
	}

	if err := h.file.Save(); err != nil {
		return wrapError("save", err)
	}
	return nil
}

func (h *Hosts) randUint32(max uint32) (uint32, error) {
	if h.randFunc == nil {
		return defaultRandFunc(max)
	}
	return h.randFunc(max)
}

// defaultRandFunc is a cryptographically secure random number generator.
func defaultRandFunc(max uint32) (uint32, error) {
	bigInt, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, wrapError("random int", err)
	}
	return uint32(bigInt.Int64()), nil
}

func wrapError(msg string, err error) error {
	// Wrap recognized errors to make them available to the caller.
	if errors.As(err, new(*fs.PathError)) || errors.Is(err, ErrHostnameInvalid) {
		return fmt.Errorf("lib127: %s: %w", msg, err)
	}

	// Make unrecognized errors opaque to limit the API surface area.
	return fmt.Errorf("lib127: %s: %v", msg, err)
}

func isLocalhost(hostname string) bool {
	return hostname == "localhost" || hostname == "localhost.localdomain"
}
