package error_handling

import (
	"fmt"
	"github.com/sevlyar/go-daemon"
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

type EnvInfo struct {
	Env           []string
	LogFileName   string
	NamedPipeName string
}

func EnvInit(name string) (info EnvInfo, err error) {
	pipeKey := strings.ToUpper(name) + "_NAMED_PIPE"
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
	baseName := fmt.Sprintf("%s-%d", name, os.Getpid())
	info.LogFileName = filepath.Join(os.TempDir(), baseName+".log")
	info.NamedPipeName = filepath.Join(os.TempDir(), baseName+".pipe")
	info.Env = []string{
		pipeKey + "=" + info.NamedPipeName,
	}
	err = syscall.Mkfifo(info.NamedPipeName, 0700)
	return
}

func EnvCleanup(info EnvInfo) {
	if daemon.WasReborn() {
		return
	}
	// the following only runs in the parent process
	err := syscall.Unlink(info.NamedPipeName)
	if err != nil {
		log.Panicln(err)
	}
}
