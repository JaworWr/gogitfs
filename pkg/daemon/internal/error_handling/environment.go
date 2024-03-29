package error_handling

import (
	"fmt"
	"github.com/sevlyar/go-daemon"
	"gogitfs/pkg/daemon/internal/environment"
	"os"
	"path/filepath"
	"syscall"
)

// EnvInfo contains information about daemon environment related to error handling.
type EnvInfo struct {
	Env           []string
	NamedPipeName string
}

const (
	pipeKey string = "_DAEMON_NAMED_PIPE"
)

// GetDaemonEnv gets environment information and initializes the named pipe.
func GetDaemonEnv() (info EnvInfo, err error) {
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
	if err != nil {
		err = fmt.Errorf("cannot create named pipe: %w", err)
	}
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
		panic("Error during cleanup: " + err.Error())
	}
}
