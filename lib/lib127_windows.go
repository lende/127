package lib127

import "os"

// Default hosts-file on Windows.
var HostsFile = os.Getenv("SystemRoot") + "\\System32\\drivers\\etc\\hosts"
