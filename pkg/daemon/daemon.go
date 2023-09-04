// Package daemon contains code allowing daemon process creation.
package daemon

import (
	"flag"
	"github.com/sevlyar/go-daemon"
	"gogitfs/pkg/daemon/environment"
	"gogitfs/pkg/daemon/internal/error_handling"
	"gogitfs/pkg/error_handler"
)

// ProcessInfo is an interface for types representing daemon processes.
type ProcessInfo interface {
	// DaemonArgs will be called in the daemon process. Then, the returned object will be used
	// to parse command line arguments provided to the child process.
	DaemonArgs() DaemonArgs
	// DaemonEnv returns extra environment variables to be set in the daemon process,
	// in the standard format "key=value".
	DaemonEnv() []string
	// DaemonProcess is the actual entry point of the daemon process. It should either
	// call errHandler (if an error occurs), which will terminate the process, or call succHandler if no error occured.
	// After callin succHandler the process is free to continue as necessary.
	// Provided args object is obtained by calling DaemonArgs() and using the result's methods to parse
	// command line arguments. As such, it is guaranteed to be of the same type as the return type of DaemonArgs().
	DaemonProcess(args DaemonArgs, errHandler error_handler.ErrorHandler, succHandler SuccessHandler)
}

// SpawnDaemon spawns the daemon process. args should be compatible with the return type of info.DaemonArgs().
// processName is used to define environment variables and file names. If the daemon process calls errHandler,
// the error will be returned by this function in the parent process. Otherwise, this function returns nil
// as soon as the child process calls succHandler.
func SpawnDaemon(args DaemonArgs, info ProcessInfo, processName string) error {
	environment.Init(processName)
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
	// parse command line args
	args := info.DaemonArgs()
	args.Setup()
	flag.Parse()
	err = args.HandlePositionalArgs(flag.Args())
	if err != nil {
		panic("argument mismatch: " + err.Error())
	}
	// run actual process code
	info.DaemonProcess(args, sender, sender)
}
