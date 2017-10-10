package lib127

import (
	"fmt"
	"net"
	"os/exec"
	"strings"

	"github.com/lende/hostsfile/lib"
)

func init() {
	_add, _get, _remove := add, get, remove

	// Add loopback alias when creating a new mapping on macOS.
	add = func(h *hostsfile.Hostsfile, hostname, ip string) error {
		if err := _add(h, hostname, ip); err != nil {
			return err
		}
		return createAlias(ip)
	}

	// Create missing aliases when getting (aliases are removed upon reboot).
	get = func(h hostsfile.Hostsfile, hostname string) (string, error) {
		ip, _ := _get(h, hostname)
		return ip, createAlias(ip)
	}

	// Remove loopback alias along with mapping.
	remove = func(h *hostsfile.Hostsfile, hostname, ip string) error {
		if err := _remove(h, hostname, ip); err != nil {
			return err
		}
		if strings.HasPrefix(ip, "127.") {
			// Do not delete alias if another hostname is mapped to it.
			for _, r := range h.Records() {
				if r.IpAddress.String() == ip {
					return nil
				}
			}
			if o, err := exec.Command("ifconfig", "lo0", "-alias", ip).CombinedOutput(); err != nil {
				return fmt.Errorf("failed to delete alias:\n\t%v", string(o))
			}
		}
		return nil
	}
}

func createAlias(ip string) error {
	if strings.HasPrefix(ip, "127.") {
		// Do not create an alias if IP address is already mapped.
		if _, err := net.LookupAddr(ip); err == nil {
			return nil
		}
		if o, err := exec.Command("ifconfig", "lo0", "alias", ip, "up").CombinedOutput(); err != nil {
			return fmt.Errorf("failed to create alias:\n\t%v", string(o))
		}
	}
	return nil
}
