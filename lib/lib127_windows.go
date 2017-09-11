package lib127

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
)

func init() {

	// Set the default hosts-file on Windows.
	HostsFile = os.Getenv("SystemRoot") + "\\System32\\drivers\\etc\\hosts"

	// Set an appropriate default backup file on Windows.
	BackupFile = HostsFile + ".127-old"

	// Create temporary file in hosts-directory to inherit correct permissions.
	_tempFile := tempFile
	tempFile = func(dir string, hosts os.FileInfo) (*os.File, error) {
		return _tempFile(filepath.Dir(HostsFile), hosts)
	}

	// Convert newlines from "\n" to "\r\n" on Windows.
	fileWriter = func(f *os.File) io.Writer { return nlWriter{f} }
}

type nlWriter struct{ io.Writer }

func (w nlWriter) Write(p []byte) (n int, err error) {
	return w.Writer.Write(bytes.Replace(p, []byte("\n"), []byte("\r\n"), -1))
}
