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

func Init(name string) {
	DaemonName = name
	DaemonParentPid = os.Getpid()

	if LogFileName == "" {
		fname := fmt.Sprintf("%s-%d.log", DaemonName, DaemonParentPid)
		LogFileName = filepath.Join(os.TempDir(), fname)
	}
}

func SetupFlags() {
	flag.StringVar(&LogFileName, "log-path", "", "log file name")
}
