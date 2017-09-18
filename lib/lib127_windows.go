package lib127

import (
	"bytes"
	"io"
	"os"
)

// Default hosts-file on Windows.
var HostsFile = os.Getenv("SystemRoot") + "\\System32\\drivers\\etc\\hosts"

func init() {
	// Convert newlines from "\n" to "\r\n" on Windows.
	fileWriter = func(f *os.File) io.Writer { return nlWriter{f} }
}

type nlWriter struct{ io.Writer }

func (w nlWriter) Write(p []byte) (n int, err error) {
	return w.Writer.Write(bytes.Replace(p, []byte("\n"), []byte("\r\n"), -1))
}
