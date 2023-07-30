package environment

import (
	"flag"
	"fmt"
	"os"
)

var DaemonName string
var DaemonParentPid int

var LogFileName string

func Init(name string) {
	DaemonName = name
	DaemonParentPid = os.Getpid()

	if LogFileName != "" {
		LogFileName = fmt.Sprintf("/tmp/%s-%d.log", DaemonName, DaemonParentPid)
	}
}

func SetupFlags() {
	flag.StringVar(&LogFileName, "log-path", "", "log file name")
}
