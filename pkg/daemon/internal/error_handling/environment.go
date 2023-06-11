package error_handling

import (
	"fmt"
	"github.com/sevlyar/go-daemon"
	"os"
)

type EnvInfo struct {
	Env         []string
	LogFileName string
}

func EnvInit(name string) (info EnvInfo) {
	if daemon.WasReborn() {
		return
	}
	// the following only runs in the parent process
	pid := os.Getpid()
	info.LogFileName = fmt.Sprintf("/tmp/%s-%d.log", name, pid)
	return
}

func EnvCleanup() {
	if daemon.WasReborn() {
		return
	}
	// the following only runs in the parent process
}
