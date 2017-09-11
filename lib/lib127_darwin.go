package lib127

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/kevinburke/hostsfile/lib"
)

func init() {
	var _add, _remove = add, remove

	// Add loopback alias when creating new mapping on macOS.
	add = func(h *hostsfile.Hostsfile, hostname, ip string) error {
		if err := _add(h, hostname, ip); err != nil {
			return err
		}
		if strings.HasPrefix(ip, "127.") {
			if o, err := exec.Command("ifconfig", "lo0", "alias", ip, "up").CombinedOutput(); err != nil {
				return fmt.Errorf("failed to create alias:\n\t%v", string(o))
			}
		}
		return nil
	}

	// Remove loopback alias along with mapping.
	remove = func(h *hostsfile.Hostsfile, hostname, ip string) error {
		if err := _remove(h, hostname, ip); err != nil {
			return err
		}
		if strings.HasPrefix(ip, "127.") {
			if o, err := exec.Command("ifconfig", "lo0", "-alias", ip).CombinedOutput(); err != nil {
				return fmt.Errorf("failed to delete alias:\n\t%v", string(o))
			}
		}
		return nil
	}
}
