package lib127_test

import (
	"errors"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/lende/127/lib127"
)

func TestOperations(t *testing.T) {
	t.Parallel()

	h := newHosts(t)
	call(h.RandomIP()).assertIP(t, "127.10.87.204")
	call(h.RandomIP()).assertIP(t, "127.174.217.245")
	call(h.RandomIP()).assertIP(t, "127.209.224.187")
	call(h.GetIP("example.test")).assertIP(t, "127.75.38.138")
	call(h.Set("example.test")).assertIP(t, "127.75.38.138")
	call(h.Remove("example.test")).assertIP(t, "127.75.38.138")
	call(h.Remove("example.test")).assertIP(t, "")
	call(h.GetIP("example.test")).assertIP(t, "")
	call(h.Set("Hello世界")).assertIP(t, "127.99.234.97")
	call(h.GetIP("Hello世界")).assertIP(t, "127.99.234.97")
	call(h.GetIP("xn--hello-ck1hg65u")).assertIP(t, "127.99.234.97")
	call(h.Remove("xn--hello-ck1hg65u")).assertIP(t, "127.99.234.97")
	call(h.GetIP("Hello世界")).assertIP(t, "")
	call(h.Set("")).assertErrorIs(t, lib127.ErrInvalidHostname)
	call(h.Remove("")).assertErrorIs(t, lib127.ErrInvalidHostname)
	call(h.Set("foo bar")).assertErrorIs(t, lib127.ErrInvalidHostname)
	call(h.Set("foo_bar")).assertErrorIs(t, lib127.ErrInvalidHostname)
	call(h.Set("192.168.0.1")).assertErrorIs(t, lib127.ErrInvalidHostname, lib127.ErrHostnameIsIP)
	call(h.Remove("localhost")).assertErrorIs(t, lib127.ErrCannotRemoveLocalhost)
}

func TestErrors(t *testing.T) {
	t.Parallel()

	var pathError *fs.PathError
	missingFile := filepath.Join(t.TempDir(), "hosts")
	h := lib127.NewHosts(missingFile)
	call(h.GetIP("example.test")).
		assertErrorIs(t, fs.ErrNotExist).
		assertErrorAs(t, &pathError)
	if path := pathError.Path; path != missingFile {
		t.Errorf("Unepexected path: %q", path)
	}

	var hostErr lib127.HostnameError
	h = newHosts(t)
	call(h.Set("foo/bar")).
		assertErrorIs(t, lib127.ErrInvalidHostname).
		assertErrorAs(t, &hostErr)
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

type output struct {
	ip  string
	err error
}

func call(ip string, err error) output {
	return output{ip: ip, err: err}
}

func (o output) assertIP(t *testing.T, ip string) output {
	t.Helper()

	o.assertErrorIs(t, nil)

	if o.ip != ip {
		t.Errorf("want IP: %q, got: %q", ip, o.ip)
	}

	return o
}

func (o output) assertErrorIs(t *testing.T, errs ...error) output {
	t.Helper()

	for _, err := range errs {
		if !errors.Is(o.err, err) {
			t.Errorf("unexpected error: %v", o.err)
		}
	}
	return o
}

func (o output) assertErrorAs(t *testing.T, vals ...any) output {
	t.Helper()

	for _, val := range vals {
		if !errors.As(o.err, val) {
			t.Errorf("unexpected error: %v", o.err)
		}
	}

	return o
}
