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
	// ErrHostnameInvalid indicates that the hostname is invalid.
	ErrInvalidHostname = hosts.ErrInvalidHostname

	// ErrHostnameIsIP indicates that the hostname is an IP address.
	ErrHostnameIsIP = hosts.ErrHostnameIsIP

	// ErrCannotRemoveLocalhost is returned by Hosts.Remove when localhost was
	// given as the hostname.
	ErrCannotRemoveLocalhost = errors.New("127: cannot remove localhost")
)

// HostnameError is implemented by hostname errors. Use errors.As to uncover
// this interface.
type HostnameError interface {
	error
	Hostname() string
}

// Hosts provide methods for mapping hostnames to random IP addresses. Its zero
// value is usable and provides good defaults.
//
// Methods may return errors that wraps HostnameError or *fs.PathError.
type Hosts struct {
	filename string
	randFunc func(uint32) (uint32, error)
}

// NewHosts creates a new Hosts using the given file. If filename is "" the
// default filename is used.
func NewHosts(filename string) *Hosts {
	return &Hosts{filename: filename}
}

// RandomIP returns a random unassigned loopback address.
func (h *Hosts) RandomIP() (string, error) {
	f, err := h.hostsFile()
	if err != nil {
		return "", err
	}
	return h.randomAvailableIP(f)
}

// Set maps the specified hostname to a random unnasigned loopback address, and
// returns that IP. If the hostname is already mapped, we return the already
// assigned IP address instead.
func (h *Hosts) Set(hostname string) (string, error) {
	if isLocalhost(hostname) {
		return "127.0.0.1", nil
	}

	f, err := h.hostsFile()
	if err != nil {
		return "", err
	}

	if ip, err := getIP(f, hostname); ip != "" || err != nil {
		return ip, err
	}

	ip, err := h.randomAvailableIP(f)
	if err != nil {
		return "", err
	}

	if err = f.Set(hostname, ip); err != nil {
		return "", wrapError("set hostname", err)
	}
	if err := save(f); err != nil {
		return "", err
	}

	return ip, nil
}

// GetIP gets the IP associated with the specified hostname. Returns the empty
// string if hostname were not found.
func (h *Hosts) GetIP(hostname string) (string, error) {
	if isLocalhost(hostname) {
		return "127.0.0.1", nil
	}

	f, err := h.hostsFile()
	if err != nil {
		return "", err
	}

	return getIP(f, hostname)
}

// Remove unmaps the specified hostname and returns the associated IP. Returns
// the empty string if hostname were not found.
func (h *Hosts) Remove(hostname string) (string, error) {
	if isLocalhost(hostname) {
		return "", ErrCannotRemoveLocalhost
	}

	f, err := h.hostsFile()
	if err != nil {
		return "", err
	}

	ip, err := getIP(f, hostname)
	if ip == "" || err != nil {
		return "", err
	}

	if err = f.Remove(hostname); err != nil {
		return "", wrapError("remove hostname", err)
	}

	if err = save(f); err != nil {
		return "", err
	}

	return ip, nil
}

func (h *Hosts) hostsFile() (*hosts.File, error) {
	if h.filename == "" {
		h.filename = DefaultHostsFile
	}
	f, err := hosts.Open(h.filename)
	if err != nil {
		return nil, wrapError("open hosts file", err)
	}
	return f, nil
}

func (h *Hosts) randUint32(max uint32) (uint32, error) {
	if h.randFunc == nil {
		return defaultRandFunc(max)
	}
	return h.randFunc(max)
}

const (
	minIP uint32 = 2130706434 // 127.0.0.2
	maxIP uint32 = 2147483646 // 127.255.255.254
)

func (h *Hosts) randomAvailableIP(f *hosts.File) (string, error) {
	ip := make(net.IP, net.IPv4len)

	for {
		// Generate a random offset.
		offset, err := h.randUint32(maxIP - minIP)
		if err != nil {
			return "", wrapError("generate random offset", err)
		}

		// Add random offset and convert integer to IP address.
		binary.BigEndian.PutUint32(ip, minIP+offset)
		if f.HasIP(ip.String()) {
			continue
		}

		return ip.String(), nil
	}
}

func getIP(f *hosts.File, hostname string) (string, error) {
	ip, err := f.GetIP(hostname)
	if err != nil {
		return "", wrapError("get hostname IP", err)
	}
	return ip, nil
}

func save(f *hosts.File) error {
	if err := f.Save(); err != nil {
		return wrapError("save changes", err)
	}
	return nil
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
	if errors.As(err, new(*fs.PathError)) ||
		errors.As(err, new(HostnameError)) {
		return fmt.Errorf("lib127: %s: %w", msg, err)
	}

	// Make unrecognized errors opaque to limit the API surface area.
	return fmt.Errorf("lib127: %s: %v", msg, err)
}

func isLocalhost(hostname string) bool {
	return hostname == "localhost" || hostname == "localhost.localdomain"
}
