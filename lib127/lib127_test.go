package lib127_test

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/lende/127/lib127"
)

func TestRandomIP(t *testing.T) {
	h := newHosts(t)

	tests := []struct{ wantIP, wantErr string }{
		{"127.134.24.251", "<nil>"},
		{"127.17.127.229", "<nil>"},
		{"127.129.197.227", "<nil>"},
	}
	for _, tt := range tests {
		ip, err := h.RandomIP()
		if ip != tt.wantIP || fmt.Sprint(err) != tt.wantErr {
			t.Errorf("randomIP()\n\tgot:  %q, %v\n\twant: %q, %v", ip, err, tt.wantIP, tt.wantErr)
		}
	}
}

func TestOperations(t *testing.T) {
	h := newHosts(t)

	steps := []struct {
		fn                            func(hostname string) (ip string, err error)
		op, hostname, wantIP, wantErr string
	}{
		{h.GetIP, "GetIP", "example.test", "127.75.38.138", "<nil>"},
		{h.Set, "Set", "example.test", "127.75.38.138", "<nil>"},
		{h.Remove, "Remove", "example.test", "127.75.38.138", "<nil>"},
		{h.Remove, "Remove", "example.test", "", "<nil>"},
		{h.GetIP, "GetIP", "example.test", "", "<nil>"},
		{h.Set, "Set", "Hello世界", "127.134.24.251", "<nil>"},
		{h.GetIP, "GetIP", "Hello世界", "127.134.24.251", "<nil>"},
		{h.GetIP, "GetIP", "xn--hello-ck1hg65u", "127.134.24.251", "<nil>"},
		{h.Remove, "Remove", "xn--hello-ck1hg65u", "127.134.24.251", "<nil>"},
		{h.GetIP, "GetIP", "Hello世界", "", "<nil>"},
		{h.Set, "Set", "foo bar", "", `hostsfile: invalid hostname "foo bar": idna: disallowed rune U+0020`},
		{h.Set, "Set", "192.168.0.1", "", `hostsfile: invalid hostname "192.168.0.1": looks like an IP address`},
		{h.Set, "Set", "foo_bar", "", `hostsfile: invalid hostname "foo_bar": idna: disallowed rune U+005F`},
	}
	for _, s := range steps {
		ip, err := s.fn(s.hostname)
		if ip != s.wantIP || fmt.Sprint(err) != s.wantErr {
			t.Errorf("%s(%#v)\n\tgot:  %q, %v\n\twant: %q, %v", s.op, s.hostname, ip, err, s.wantIP, s.wantErr)
		}
	}
}

func newHosts(t *testing.T) *lib127.Hosts {
	data := `127.0.0.1 localhost localhost.localdomain
127.0.0.2 localhost2
127.75.38.138 example.test
`
	hostsFile := filepath.Join(t.TempDir(), "hosts")
	if err := os.WriteFile(hostsFile, []byte(data), 0600); err != nil {
		t.Errorf("Unexpected error:\n\t%v", err)
	}

	h := lib127.NewHosts(hostsFile)

	// Ensure predictable results with a pseudo-random number generator.
	r := rand.New(rand.NewSource(1))
	h.SetRandFunc(func(max uint32) (uint32, error) {
		return uint32(r.Int63n(int64(max))), nil
	})

	return h
}
