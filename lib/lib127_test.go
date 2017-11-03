package lib127

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"testing"

	"github.com/lende/127/internal/hostsfile"
)

func TestRandomIP(t *testing.T) {
	defer removeFile(tmpHosts(t))
	hosts, err := hostsfile.Open(HostsFile)
	if err != nil {
		t.Errorf("hostsfile.Open(HostsFile):\nUnexpected error:\n\t%v", err)
	}
	tests := []struct{ block, err string }{
		{"127.0.0.0/33", "invalid CIDR address: 127.0.0.0/33"},
		{"127.0.0.0/32", "address block too small: 127.0.0.0/32"},
		{"127.0.0.0/31", "address block too small: 127.0.0.0/31"},
		{"127.0.0.0/30", "no unnasigned IPs in address block: 127.0.0.0/30"},
		{"127.0.0.0/29", "<nil>"},
		{"127.0.0.0/8", "<nil>"},
	}
	for _, test := range tests {
		_, err := randomIP(hosts, test.block)
		if fmt.Sprint(err) != test.err {
			if test.err == "<nil>" {
				t.Errorf("RandomIP():\nUnexpected error:\n\t%v", err)
			} else {
				t.Errorf("RandomIP():\nExpected error:\n\t%v\ngot:\n\t%v", test.err, err)
			}
		}
	}
}

func TestOperations(t *testing.T) {
	defer removeFile(tmpHosts(t))
	// Seed the random number generator with a fixed value so randomIP produce
	// predictable results.
	rand.Seed(1)
	steps := []struct {
		fn                    func(hostname string) (ip string, err error)
		op, hostname, ip, err string
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
		{Set, "Set", "foo bar", "", "invalid hostname: foo bar"},
		{Set, "Set", "192.168.0.1", "", "invalid hostname: 192.168.0.1"},
		{Set, "Set", "foo_bar", "", "invalid hostname: foo_bar"},
	}
	for _, test := range steps {
		if ip, err := test.fn(test.hostname); fmt.Sprint(err) != test.err {
			if test.err == "<nil>" {
				t.Errorf("%v(%#v):\nUnexpected error:\n\t%v", test.op, test.hostname, err)
			} else {
				t.Errorf("%v(%#v):\nExpected error:\n\t%v\ngot:\n\t%v", test.op, test.hostname, test.err, err)
			}
		} else if ip != test.ip {
			t.Errorf("%v(%#v):\nWant return value:\n\t%#v,\nhave:\n\t%#v", test.op, test.hostname, test.ip, ip)
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

func tmpHosts(t *testing.T) *os.File {
	var hosts = `127.0.0.1 localhost localhost.localdomain
127.0.0.2 localhost2
127.75.38.138 example.test
`
	tmp, err := ioutil.TempFile("", "hosts.127-test")
	if err != nil {
		t.Errorf("Unexpected error:\n\t%v", err)
	}
	if err := ioutil.WriteFile(tmp.Name(), []byte(hosts), 0644); err != nil {
		t.Errorf("Unexpected error:\n\t%v", err)
	}
	HostsFile = tmp.Name()
	return tmp
}

func removeFile(f *os.File) {
	f.Close()
	os.Remove(f.Name())
}
