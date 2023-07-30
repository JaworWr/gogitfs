package daemon

import (
	"flag"
	"github.com/sevlyar/go-daemon"
	"gogitfs/pkg/daemon/environment"
	"gogitfs/pkg/daemon/internal/error_handling"
	"gogitfs/pkg/error_handler"
)

type ProcessInfo interface {
	DaemonArgs() DaemonArgs
	DaemonEnv() []string
	DaemonProcess(args DaemonArgs, errHandler error_handler.ErrorHandler, succHandler SuccessHandler)
}

func SpawnDaemon(args DaemonArgs, info ProcessInfo, name string) error {
	environment.Init(name)
	envInfo, err := error_handling.GetDaemonEnv()
	if err != nil {
		panic("Cannot setup daemon environment.\nError: " + err.Error())
	}
	defer error_handling.CleanupDeamonEnv(envInfo)

	env := append(envInfo.Env, info.DaemonEnv()...)
	ctx := daemon.Context{
		Args:        argsToFullList(args),
		Env:         env,
		LogFileName: environment.LogFileName,
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
	args := info.DaemonArgs()
	args.Setup()
	flag.Parse()
	err = args.HandlePositionalArgs(flag.Args())
	if err != nil {
		panic("argument mismatch: " + err.Error())
	}
	info.DaemonProcess(args, sender, sender)
}
