package error_handling

import (
	"fmt"
	"github.com/sevlyar/go-daemon"
	"gogitfs/pkg/daemon/environment"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// EnvInfo contains information about daemon environment related to error handling.
type EnvInfo struct {
	Env           []string
	NamedPipeName string
}

// GetDaemonEnv gets environment information and initializes the named pipe.
func GetDaemonEnv() (info EnvInfo, err error) {
	// key under which the named pipe's name appears in the environment
	pipeKey := strings.ToUpper(environment.DaemonName) + "_NAMED_PIPE"
	if daemon.WasReborn() {
		// this runs in the child process
		pipeName, ok := os.LookupEnv(pipeKey)
		if !ok {
			err = fmt.Errorf("missing environment variable: %v", pipeKey)
			return
		}
		info.NamedPipeName = pipeName
		return
	}
	// the following only runs in the parent process
	// make the named pipe
	baseName := fmt.Sprintf("%s-%d", environment.DaemonName, environment.DaemonParentPid)
	info.NamedPipeName = filepath.Join(os.TempDir(), baseName+".pipe")
	info.Env = []string{
		pipeKey + "=" + info.NamedPipeName,
	}
	err = syscall.Mkfifo(info.NamedPipeName, 0700)
	return
}

func CleanupDeamonEnv(info EnvInfo) {
	if daemon.WasReborn() {
		return
	}
	// the following only runs in the parent process
	// delete the named pipe
	err := syscall.Unlink(info.NamedPipeName)
	if err != nil {
		panic("Error during cleanup:" + err.Error())
	}
}
