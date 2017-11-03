package hostsfile

import (
	"fmt"
	"net"
	"os"

	"github.com/kevinburke/hostsfile/lib"

	"golang.org/x/net/idna"
)

// Location is the OS-specific hosts-file location.
var Location = hostsfile.Location

// Record represent a single line from a hosts-file.
type Record hostsfile.Record

// Hostsfile is an in-memory representation of a hosts-file.
type Hostsfile struct {
	hostsfile hostsfile.Hostsfile
	filename  string
}

// Open opens the hosts-file and returns a representation.
func Open(filename string) (*Hostsfile, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	h, err := hostsfile.Decode(f)
	if err != nil {
		return nil, err
	}
	return &Hostsfile{h, filename}, nil
}

// GetIP returns the IP address associated with the given hostname, if any.
func (h Hostsfile) GetIP(hostname string) (ip string, err error) {
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
func (h Hostsfile) Records() (rs []Record) {
	for _, r := range h.hostsfile.Records() {
		rs = append(rs, Record(*r))
	}
	return rs
}

// Set maps the specified hostname to the given IP.
func (h *Hostsfile) Set(hostname, ip string) (err error) {
	if hostname, err = adaptHostname(hostname); err != nil {
		return err
	}
	return h.hostsfile.Set(net.IPAddr{IP: net.ParseIP(ip)}, hostname)
}

// Remove removes the given hostname mapping.
func (h *Hostsfile) Remove(hostname string) (err error) {
	if hostname, err = adaptHostname(hostname); err != nil {
		return err
	}
	h.hostsfile.Remove(hostname)
	return nil
}

// Save saves the changes to the hosts-file.
func (h Hostsfile) Save() error {
	f, err := os.OpenFile(h.filename, os.O_WRONLY|os.O_TRUNC, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	return hostsfile.Encode(f, h.hostsfile)
}

// adaptHostname validates the given hostname and converts it from unicode to
// IDNA Punycode.
func adaptHostname(hostname string) (string, error) {
	h, err := idna.Lookup.ToASCII(hostname)
	if err != nil || net.ParseIP(hostname) != nil {
		return "", fmt.Errorf("invalid hostname: %v", hostname)
	}
	return h, nil
}
