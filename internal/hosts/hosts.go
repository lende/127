package hosts

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"

	hostsfile "github.com/kevinburke/hostsfile/lib"
	"golang.org/x/net/idna"
)

// ErrHostnameInvalid is returned when the given hostname is invalid.
var ErrInvalidHostname = errors.New("hosts: invalid hostname")

// FileLocation is the OS-specific hosts-file location.
var FileLocation = hostsfile.Location

// Record represent a single line from a hosts-file.
type Record hostsfile.Record

// File is an in-memory representation of a hosts-file.
type File struct {
	hostsfile hostsfile.Hostsfile
	filename  string
}

// Open opens the hosts-file and returns a representation.
func Open(filename string) (*File, error) {
	f, err := os.Open(filepath.Clean(filename))
	if err != nil {
		return nil, fmt.Errorf("hosts: open file: %w", err)
	}
	defer f.Close()

	h, err := hostsfile.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("hosts: decode file: %v", err)
	}
	return &File{h, filename}, nil
}

// HasIP returns true if the ip exists in the hosts file.
func (h File) HasIP(ip string) bool {
	for _, r := range h.Records() {
		if r.IpAddress.IP.String() == ip {
			return true
		}
	}
	return false
}

// GetIP returns the IP address associated with the given hostname, if any.
func (h File) GetIP(hostname string) (ip string, err error) {
	if hostname, err = adaptHostname(hostname); err != nil {
		return "", err
	}
	for _, r := range h.hostsfile.Records() {
		if r.Hostnames[hostname] {
			return r.IpAddress.String(), nil
		}
	}
	return "", nil
}

// Records returns an array of all entries in the hosts-file.
func (h File) Records() (rs []*Record) {
	for _, r := range h.hostsfile.Records() {
		if len(r.Hostnames) == 0 {
			continue
		}

		rs = append(rs, (*Record)(r))
	}
	return rs
}

// Set maps the specified hostname to the given IP.
func (h *File) Set(hostname, ip string) (err error) {
	if hostname, err = adaptHostname(hostname); err != nil {
		return err
	}
	if err := h.hostsfile.Set(net.IPAddr{IP: net.ParseIP(ip)}, hostname); err != nil {
		return fmt.Errorf("hosts: set hostname: %v", err)
	}
	return nil
}

// Remove removes the given hostname mapping.
func (h *File) Remove(hostname string) (err error) {
	if hostname, err = adaptHostname(hostname); err != nil {
		return err
	}
	h.hostsfile.Remove(hostname)
	return nil
}

// Save saves the changes to the hosts-file.
func (h File) Save() error {
	f, err := os.OpenFile(h.filename, os.O_WRONLY|os.O_TRUNC, 0)
	if err != nil {
		return fmt.Errorf("hosts: open file: %w", err)
	}

	if err := hostsfile.Encode(f, h.hostsfile); err != nil {
		_ = f.Close()
		return fmt.Errorf("hosts: encode file: %v", err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("hosts: close file: %w", err)
	}

	return nil
}

type hostnameError struct {
	format   string
	hostname string
	err      error
}

func (e hostnameError) Hostname() string {
	return e.hostname
}

func (e hostnameError) Error() string {
	msg := fmt.Sprintf(e.format, e.hostname)
	if e.err != nil {
		msg += ": " + e.err.Error()
	}
	return msg
}

func (e hostnameError) Is(err error) bool {
	return err == ErrInvalidHostname
}

// adaptHostname validates the given hostname and converts it from unicode to
// IDNA Punycode.
func adaptHostname(hostname string) (string, error) {
	if net.ParseIP(hostname) != nil {
		return "", hostnameError{
			format:   "hosts: adapt hostname %q: host is IP address",
			hostname: hostname,
		}
	}
	h, err := idna.Lookup.ToASCII(hostname)
	if err != nil {
		return "", hostnameError{
			format:   "hosts: adapt hostname %q",
			hostname: hostname,
			err:      err,
		}
	}
	return h, nil
}
