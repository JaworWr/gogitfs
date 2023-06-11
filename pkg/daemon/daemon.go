package daemon

import (
	"github.com/sevlyar/go-daemon"
	"gogitfs/pkg/daemon/internal/error_handling"
	"gogitfs/pkg/error_handler"
	"os"
)

type ProcessInfo interface {
	DaemonArgs(args []string) []string
	DaemonEnv(env []string) []string
	DaemonProcess(errHandler error_handler.ErrorHandler, succHandler SuccessHandler)
}

type dummySuccHandler struct{}

func (h dummySuccHandler) HandleSuccess() {

}

func SpawnDaemon(info ProcessInfo, name string) error {
	envInfo, err := error_handling.EnvInit(name)
	if err != nil {
		return err
	}
	defer error_handling.EnvCleanup(envInfo)

	args := info.DaemonArgs(os.Args)
	env := append(envInfo.Env, info.DaemonEnv(os.Environ())...)
	ctx := daemon.Context{
		Args:        args,
		Env:         env,
		LogFileName: envInfo.LogFileName,
		LogFilePerm: 0755,
	}
	child, err := ctx.Reborn()
	if err != nil {
		return err
	}

	if child == nil {
		// child code - run actual process
		errHandler := error_handler.FatalHandler
		succHandler := dummySuccHandler{}
		info.DaemonProcess(errHandler, succHandler)
		return nil
	}
	// parent code - handle errors from child
	err = parentProcessPostSpawn()
	return err
}

func parentProcessPostSpawn() (err error) {
	// handle errors from child - for now, do nothing
	return
}
