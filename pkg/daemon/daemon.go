package daemon

import (
	"github.com/sevlyar/go-daemon"
	"gogitfs/pkg/daemon/internal/error_handling"
	"gogitfs/pkg/error_handler"
	"os"
)

type ProcessInfo interface {
	DaemonArgs(args []string) []string
	DaemonEnv() []string
	DaemonProcess(errHandler error_handler.ErrorHandler, succHandler SuccessHandler)
}

func SpawnDaemon(info ProcessInfo, name string) error {
	envInfo, err := error_handling.EnvInit(name)
	if err != nil {
		return err
	}
	defer error_handling.EnvCleanup(envInfo)

	args := info.DaemonArgs(os.Args)
	env := append(envInfo.Env, info.DaemonEnv()...)
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
		childProcessPostSpawn(info, envInfo)
		return nil
	}
	// parent code - handle errors from child
	err = parentProcessPostSpawn(envInfo)
	return err
}

func parentProcessPostSpawn(envInfo error_handling.EnvInfo) error {
	receiver, err := error_handling.NewSubprocessErrorReceiver(envInfo.NamedPipeName)
	if err != nil {
		panic("Unable to setup daemon error receiver\nError: " + err.Error())
	}
	defer receiver.Close()
	return receiver.Receive()
}

func childProcessPostSpawn(info ProcessInfo, envInfo error_handling.EnvInfo) {
	sender, err := error_handling.NewSubprocessErrorSender(envInfo.NamedPipeName)
	if err != nil {
		panic("Unable to setup daemon error sender\nError: " + err.Error())
	}
	defer sender.Close()
	info.DaemonProcess(sender, sender)
}
