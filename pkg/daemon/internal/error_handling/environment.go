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

var pipeKey string

type EnvInfo struct {
	Env           []string
	LogFileName   string
	NamedPipeName string
}

func EnvInit(name string) (info EnvInfo, err error) {
	initKeys(name)
	if daemon.WasReborn() {
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

func EnvCleanup(info *EnvInfo) {
	if daemon.WasReborn() {
		return
	}
	// the following only runs in the parent process
	err := syscall.Unlink(info.NamedPipeName)
	if err != nil {
		log.Panicln(err)
	}
}

func initKeys(name string) {
	pipeKey = strings.ToUpper(name) + "_NAMED_PIPE"
}
