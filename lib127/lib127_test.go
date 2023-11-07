package lib127_test

import (
	"errors"
	"io/fs"
	"math/rand"
	"path/filepath"
	"testing"

	"github.com/lende/127/internal/testdata"
	"github.com/lende/127/lib127"
)

func TestOperations(t *testing.T) {
	t.Parallel()

	h := newHosts(t)

	// Pseudo-random IPs.
	call(h.RandomIP()).assertIP(t, pseudoRndIP1)
	call(h.RandomIP()).assertIP(t, pseudoRndIP2)
	call(h.RandomIP()).assertIP(t, pseudoRndIP3)

	// Various operations on the same hostname.
	const loopbackHostname, loopbackIP = "loopback.test", "127.0.0.3"
	call(h.GetIP(loopbackHostname)).assertIP(t, loopbackIP)
	call(h.Set(loopbackHostname)).assertIP(t, loopbackIP)
	call(h.Remove(loopbackHostname)).assertIP(t, loopbackIP)
	call(h.Remove(loopbackHostname)).assertIP(t, "")
	call(h.GetIP(loopbackHostname)).assertIP(t, "")
	call(h.Set(loopbackHostname)).assertIP(t, pseudoRndIP4)

	// Pre-assigned public domain.
	call(h.GetIP("example.com")).assertIP(t, "93.184.216.34")

	// Commented out hostname.
	call(h.GetIP("private.test")).assertIP(t, "")

	// Internationalized hostname.
	const chineseHostname, chinesePunicode = "Hello世界", "xn--hello-ck1hg65u"
	call(h.Set(chineseHostname)).assertIP(t, pseudoRndIP5)
	call(h.GetIP(chineseHostname)).assertIP(t, pseudoRndIP5)
	call(h.GetIP(chinesePunicode)).assertIP(t, pseudoRndIP5)
	call(h.Remove(chinesePunicode)).assertIP(t, pseudoRndIP5)
	call(h.GetIP(chineseHostname)).assertIP(t, "")

	// Various illegal arguments.
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

const (
	pseudoRndIP1 = "127.10.87.204"
	pseudoRndIP2 = "127.174.217.245"
	pseudoRndIP3 = "127.209.224.187"
	pseudoRndIP4 = "127.99.234.97"
	pseudoRndIP5 = "127.224.77.159"
)

func newHosts(t *testing.T) *lib127.Hosts {
	h := lib127.NewHosts(testdata.HostsFile(t))

	// Ensure predictable results with a pseudo-random number generator.
	r := rand.New(rand.NewSource(1))
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
