package lib127

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"path/filepath"
	"testing"
)

func TestRandomIP(t *testing.T) {
	setupTests(t)

	tests := []struct{ block, wantIP, wantErr string }{
		{"127.0.0.0/33", "", "lib127: could not parse address block: invalid CIDR address: 127.0.0.0/33"},
		{"127.0.0.0/32", "", "lib127: address block too small: 127.0.0.0/32"},
		{"127.0.0.0/31", "", "lib127: address block too small: 127.0.0.0/31"},
		{"127.0.0.0/30", "", "lib127: no unnasigned IPs in address block: 127.0.0.0/30"},
		{"127.0.0.0/29", "127.0.0.3", "<nil>"},
		{"127.0.0.0/29", "127.0.0.4", "<nil>"},
		{"127.0.0.0/8", "127.61.139.76", "<nil>"},
		{"127.0.0.0/8", "127.167.80.0", "<nil>"},
	}

	for _, tt := range tests {
		AddressBlock = tt.block
		ip, err := RandomIP()
		if ip != tt.wantIP || fmt.Sprint(err) != tt.wantErr {
			t.Errorf("randomIP(%#v)\n\tgot:  %q, %v\n\twant: %q, %v", tt.block, ip, err, tt.wantIP, tt.wantErr)
		}
	}
}

func TestOperations(t *testing.T) {
	setupTests(t)

	steps := []struct {
		fn                            func(hostname string) (ip string, err error)
		op, hostname, wantIP, wantErr string
	}{
		{GetIP, "GetIP", "example.test", "127.75.38.138", "<nil>"},
		{Set, "Set", "example.test", "127.75.38.138", "<nil>"},
		{Remove, "Remove", "example.test", "127.75.38.138", "<nil>"},
		{Remove, "Remove", "example.test", "", "<nil>"},
		{GetIP, "GetIP", "example.test", "", "<nil>"},
		{Set, "Set", "Hello世界", "127.134.24.251", "<nil>"},
		{GetIP, "GetIP", "Hello世界", "127.134.24.251", "<nil>"},
		{GetIP, "GetIP", "xn--hello-ck1hg65u", "127.134.24.251", "<nil>"},
		{Remove, "Remove", "xn--hello-ck1hg65u", "127.134.24.251", "<nil>"},
		{GetIP, "GetIP", "Hello世界", "", "<nil>"},
		{Set, "Set", "foo bar", "", `hostsfile: invalid hostname "foo bar": idna: disallowed rune U+0020`},
		{Set, "Set", "192.168.0.1", "", `hostsfile: invalid hostname "192.168.0.1": looks like an IP address`},
		{Set, "Set", "foo_bar", "", `hostsfile: invalid hostname "foo_bar": idna: disallowed rune U+005F`},
	}

	for _, s := range steps {
		ip, err := s.fn(s.hostname)
		if ip != s.wantIP || fmt.Sprint(err) != s.wantErr {
			t.Errorf("%s(%#v)\n\tgot:  %q, %v\n\twant: %q, %v", s.op, s.hostname, ip, err, s.wantIP, s.wantErr)
		}
	}
}

func TestBlockRange(t *testing.T) {
	tests := []struct{ block, minIP, maxIP string }{
		{"127.0.0.0/8", "127.0.0.1", "127.255.255.254"},
		{"127.0.0.0/24", "127.0.0.1", "127.0.0.254"},
		{"127.0.0.0/4", "112.0.0.1", "127.255.255.254"},
		{"127.0.0.24/8", "127.0.0.1", "127.255.255.254"},
		{"203.0.113.16/28", "203.0.113.17", "203.0.113.30"},
	}
	minIP, maxIP := make(net.IP, 4), make(net.IP, 4)
	for _, test := range tests {
		_, ipnet, err := net.ParseCIDR(test.block)
		if err != nil {
			t.Errorf("Unexpected error:\n\t%v", err)
		}
		minU32, maxU32 := ipSpan(ipnet)
		binary.BigEndian.PutUint32(minIP, minU32)
		binary.BigEndian.PutUint32(maxIP, maxU32)
		if test.minIP != minIP.String() || test.maxIP != maxIP.String() {
			t.Errorf("Host address range in block %v:\nwant:\n\t%v - %v,\nhave:\n\t%v - %v",
				test.block, test.minIP, test.maxIP, minIP.String(), maxIP.String())
		}
	}
}

func setupTests(t *testing.T) {
	// Ensure predictable results with a pseudo-random number generator.
	r := rand.New(rand.NewSource(1))
	randUint32 = func(max uint32) (uint32, error) {
		return uint32(r.Int63n(int64(max))), nil
	}

	data := `127.0.0.1 localhost localhost.localdomain
127.0.0.2 localhost2
127.75.38.138 example.test
`
	HostsFile = filepath.Join(t.TempDir(), "hosts")

	if err := ioutil.WriteFile(HostsFile, []byte(data), 0600); err != nil {
		t.Errorf("Unexpected error:\n\t%v", err)
	}
}
