package lib127

import (
	"encoding/binary"
	"net"
	"testing"
)

func TestConstants(t *testing.T) {
	ip := make(net.IP, 4)

	binary.BigEndian.PutUint32(ip, minIP)
	if ip.String() != "127.0.0.2" {
		t.Errorf("minIP: unexpected constant value: %d (%s)", minIP, ip)
	}

	binary.BigEndian.PutUint32(ip, maxIP)
	if ip.String() != "127.255.255.254" {
		t.Errorf("maxIP: unexpected constant value: %d (%s)", maxIP, ip)
	}
}
