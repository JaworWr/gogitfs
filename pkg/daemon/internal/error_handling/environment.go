package error_handling

import (
	"fmt"
	"github.com/sevlyar/go-daemon"
	"os"
	"path/filepath"
	"strings"
)

var pipeKey string

type EnvInfo struct {
	Env           []string
	LogFileName   string
	NamedPipeName string
}

func EnvInit(name string) (info EnvInfo) {
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
	return
}

func EnvCleanup() {
	if daemon.WasReborn() {
		return
	}
	// the following only runs in the parent process
}

func initKeys(name string) {
	pipeKey = strings.ToUpper(name) + "_NAMED_PIPE"
}
