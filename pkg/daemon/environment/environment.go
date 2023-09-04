// Package environment describes environment used by the daemon.
package environment

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

var DaemonName string
var DaemonParentPid int

var LogFileName string

// Init initializes global variables defined in this package, setting daemon name according to args
// and daemon parent PID by taking current process' PID.
func Init(daemonName string) {
	DaemonName = daemonName
	DaemonParentPid = os.Getpid()

	if LogFileName == "" {
		fname := fmt.Sprintf("%s-%d.log", DaemonName, DaemonParentPid)
		LogFileName = filepath.Join(os.TempDir(), fname)
	}
}

// SetupFlags adds necessary command-line flags.
func SetupFlags() {
	flag.StringVar(&LogFileName, "log-path", "", "log file name")
}
