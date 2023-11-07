package testdata

import (
	_ "embed"
	"os"
	"path/filepath"
	"testing"
)

//go:embed hosts
var hosts []byte

// HostsFile creates a hosts file in a temporary directory and returns the path.
func HostsFile(t *testing.T) string {
	path := filepath.Join(t.TempDir(), "hosts")
	if err := os.WriteFile(path, hosts, 0o600); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	return path
}
