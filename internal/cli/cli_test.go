package cli_test

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/lende/127/internal/cli"
)

func TestApp(t *testing.T) {
	t.Parallel()

	const hostsData = "127.0.0.1 localhost localhost.localdomain"

	err := os.WriteFile(filepath.Join(t.TempDir(), "hosts"), []byte(hostsData), 0o600)
	if err != nil {
		t.Fatalf("Writing file: %v", err)
	}

	run("localhost").assertStdout(t, "127.0.0.1")
	run("-v").assertStdout(t, "127 test-version %s/%s", runtime.GOOS, runtime.GOARCH)
	run("127.205.131.186").assertStderr(t, `127: invalid hostname: 127.205.131.186.`)
	run("-f", "non-existent").assertStderr(t, `127: open non-existent: no such file or directory.`)
}

type output struct {
	status         int
	stdout, stderr string
}

func run(args ...string) output {
	var stdout, stderr strings.Builder
	app := cli.NewApp("test-version", &stdout, &stderr)

	return output{
		status: app.Run(args...),
		stdout: stdout.String(),
		stderr: stderr.String(),
	}
}

func (o output) assertStdout(t *testing.T, format string, a ...any) {
	t.Helper()

	o.assert(t, 0, fmt.Sprintf(format, a...), "")
}

func (o output) assertStderr(t *testing.T, format string, a ...any) {
	t.Helper()

	o.assert(t, 1, "", fmt.Sprintf(format, a...))
}

func (o output) assert(t *testing.T, status int, stdout, stderr string) {
	t.Helper()

	if o.status != status {
		t.Errorf("Want status code: %d, got: %d.", status, o.status)
	}
	if got := strings.TrimSpace(o.stdout); got != stdout {
		t.Errorf("Want stdout: %q, got: %q.", stdout, got)
	}
	if got := strings.TrimSpace(o.stderr); got != stderr {
		t.Errorf("Want stderr: %q got: %q.", stderr, got)
	}
}
