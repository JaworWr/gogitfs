package environment

import "os"

var DaemonName string
var DaemonParentPid int

func Init(name string) {
	DaemonName = name
	DaemonParentPid = os.Getpid()
}
