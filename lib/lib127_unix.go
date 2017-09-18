// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

package lib127

import (
	"os"
	"syscall"
)

// Default file locations on Unix.
var (
	HostsFile  = "/etc/hosts"
	BackupFile = HostsFile + ".127-old"
)

func init() {
	// Create temporary file and copy attributes from hosts-file.
	_tempFile := tempFile
	tempFile = func(dir string, hosts os.FileInfo) (f *os.File, err error) {
		if f, err = _tempFile(dir, hosts); err != nil {
			return nil, err
		} else if err = os.Chmod(f.Name(), hosts.Mode()); err != nil {
			return nil, err
		}
		uid, gid := hosts.Sys().(*syscall.Stat_t).Uid, hosts.Sys().(*syscall.Stat_t).Gid
		if err = os.Chown(f.Name(), int(uid), int(gid)); err != nil {
			return nil, err
		}
		return f, err
	}
}
