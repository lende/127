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

	h := openHosts(t)

	// Pseudo-random IPs.
	call(h.RandomIP()).assertIP(t, pseudoRndIP1)
	call(h.RandomIP()).assertIP(t, pseudoRndIP2)
	call(h.RandomIP()).assertIP(t, pseudoRndIP3)

	// Various operations on the same hostname.
	const loopbackHostname, loopbackIP = "loopback.test", "127.0.0.3"
	call(h.IP(loopbackHostname)).assertIP(t, loopbackIP)
	call(h.Map(loopbackHostname)).assertIP(t, loopbackIP)
	call(h.Unmap(loopbackHostname)).assertIP(t, loopbackIP)
	call(h.Unmap(loopbackHostname)).assertIP(t, "")
	call(h.IP(loopbackHostname)).assertIP(t, "")
	call(h.Map(loopbackHostname)).assertIP(t, pseudoRndIP4)

	// Pre-assigned public domain.
	call(h.IP("example.com")).assertIP(t, "93.184.216.34")

	// Commented out hostname.
	call(h.IP("private.test")).assertIP(t, "")

	// Internationalized hostname.
	const chineseHostname, chinesePunicode = "Hello世界", "xn--hello-ck1hg65u"
	call(h.Map(chineseHostname)).assertIP(t, pseudoRndIP5)
	call(h.IP(chineseHostname)).assertIP(t, pseudoRndIP5)
	call(h.IP(chinesePunicode)).assertIP(t, pseudoRndIP5)
	call(h.Unmap(chinesePunicode)).assertIP(t, pseudoRndIP5)
	call(h.IP(chineseHostname)).assertIP(t, "")

	// Various illegal arguments.
	call(h.Map("")).assertErrorIs(t, lib127.ErrHostnameInvalid)
	call(h.Unmap("")).assertErrorIs(t, lib127.ErrHostnameInvalid)
	call(h.Map("foo bar")).assertErrorIs(t, lib127.ErrHostnameInvalid)
	call(h.Map("192.168.0.1")).assertErrorIs(t, lib127.ErrHostnameInvalid, lib127.ErrHostnameIsIP)
	call(h.Unmap("localhost")).assertErrorIs(t, lib127.ErrCannotUnmapLocalhost)
}

func TestFSError(t *testing.T) {
	t.Parallel()

	var pathError *fs.PathError
	missingFile := filepath.Join(t.TempDir(), "hosts")
	_, err := lib127.Open(missingFile)
	output{err: err}.
		assertErrorIs(t, fs.ErrNotExist).
		assertErrorAs(t, &pathError)
	if path := pathError.Path; path != missingFile {
		t.Errorf("Unepexected path: %q", path)
	}
}

const (
	pseudoRndIP1 = "127.10.87.204"
	pseudoRndIP2 = "127.174.217.245"
	pseudoRndIP3 = "127.209.224.187"
	pseudoRndIP4 = "127.99.234.97"
	pseudoRndIP5 = "127.224.77.159"
)

func openHosts(t *testing.T) *lib127.Hosts {
	h, err := lib127.Open(testdata.HostsFile(t))
	requireNoError(t, err)

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

	requireNoError(t, o.err)

	if o.ip != ip {
		t.Errorf("want IP: %q, got: %q", ip, o.ip)
	}

	return o
}

func (o output) assertErrorIs(t *testing.T, errs ...error) output {
	t.Helper()

	for _, err := range errs {
		if !errors.Is(o.err, err) {
			t.Errorf("error does not match %T", err)
		}
	}
	return o
}

func (o output) assertErrorAs(t *testing.T, vals ...any) output {
	t.Helper()

	for _, val := range vals {
		if !errors.As(o.err, val) {
			t.Errorf("error can not be read into %T", val)
		}
	}

	return o
}

func requireNoError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
