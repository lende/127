package lib127_test

import (
	"errors"
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/lende/127/lib127"
)

func TestRandomIP(t *testing.T) {
	t.Parallel()

	h := newHosts(t)

	tests := []struct{ wantIP, wantErr string }{
		{"127.10.87.204", "<nil>"},
		{"127.174.217.245", "<nil>"},
		{"127.209.224.187", "<nil>"},
	}
	for _, tt := range tests {
		ip, err := h.RandomIP()
		if ip != tt.wantIP || fmt.Sprint(err) != tt.wantErr {
			t.Errorf("randomIP()\n\tgot:  %q, %v\n\twant: %q, %v", ip, err, tt.wantIP, tt.wantErr)
		}
	}
}

func TestOperations(t *testing.T) {
	t.Parallel()

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
		{h.Set, "Set", "Hello世界", "127.10.87.204", "<nil>"},
		{h.GetIP, "GetIP", "Hello世界", "127.10.87.204", "<nil>"},
		{h.GetIP, "GetIP", "xn--hello-ck1hg65u", "127.10.87.204", "<nil>"},
		{h.Remove, "Remove", "xn--hello-ck1hg65u", "127.10.87.204", "<nil>"},
		{h.GetIP, "GetIP", "Hello世界", "", "<nil>"},
		{h.Set, "Set", "foo bar", "", `lib127: get hostname IP: hosts: adapt hostname "foo bar": idna: disallowed rune U+0020`},
		{h.Set, "Set", "192.168.0.1", "", `lib127: get hostname IP: hosts: adapt hostname "192.168.0.1": host is IP address`},
		{h.Set, "Set", "foo_bar", "", `lib127: get hostname IP: hosts: adapt hostname "foo_bar": idna: disallowed rune U+005F`},
	}
	for _, s := range steps {
		ip, err := s.fn(s.hostname)
		if ip != s.wantIP || fmt.Sprint(err) != s.wantErr {
			t.Errorf("%s(%#v)\n\tgot:  %q, %v\n\twant: %q, %v", s.op, s.hostname, ip, err, s.wantIP, s.wantErr)
		}
	}
}

func TestErrors(t *testing.T) {
	t.Parallel()

	h := lib127.NewHosts("testdata/no-such-file")
	_, err := h.GetIP("localhost")
	assertErrorIs(t, err, fs.ErrNotExist)

	h = newHosts(t)
	_, err = h.Set("foo/bar")
	assertErrorIs(t, err, lib127.ErrInvalidHostname)

	var hostErr lib127.HostnameError
	assertErrorAs(t, err, &hostErr)
	if host := hostErr.Hostname(); host != "foo/bar" {
		t.Errorf("Unepexected hostname: %q", host)
	}
}

func newHosts(t *testing.T) *lib127.Hosts {
	data := `127.0.0.1 localhost localhost.localdomain
127.0.0.2 localhost2
127.75.38.138 example.test
`
	hostsFile := filepath.Join(t.TempDir(), "hosts")
	if err := os.WriteFile(hostsFile, []byte(data), 0o600); err != nil {
		t.Fatalf("Unexpected error:\n\t%v", err)
	}

	h := lib127.NewHosts(hostsFile)

	// Ensure predictable results with a pseudo-random number generator.
	r := rand.New(rand.NewSource(1)) //nolint: gosec // G404: Use of weak random number generator.
	h.SetRandFunc(func(max uint32) (uint32, error) {
		return uint32(r.Int63n(int64(max))), nil
	})

	return h
}

func assertErrorIs(t *testing.T, err, target error) {
	t.Helper()

	if !errors.Is(err, target) {
		t.Errorf("Unexpected error: %q", err)
	}
}

func assertErrorAs(t *testing.T, err error, target any) {
	t.Helper()

	if !errors.As(err, target) {
		t.Errorf("Unexpected error: %q", err)
	}
}
