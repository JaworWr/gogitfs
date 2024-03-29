// Package daemon contains code allowing daemon process creation.
package daemon

import (
	"fmt"
	"github.com/sevlyar/go-daemon"
	"gogitfs/pkg/daemon/internal/environment"
	"gogitfs/pkg/daemon/internal/error_handling"
	"gogitfs/pkg/error_handler"
)

// Daemon is an interface for types representing daemon processes.
type Daemon interface {
	SerializableCliArgs
	// DaemonMain is the actual entry point of the daemon process. It should either
	// call errHandler (if an error occurs), which will terminate the process, or call succHandler if no error occured.
	// After calling succHandler the process is free to continue as necessary.
	DaemonMain(errHandler error_handler.ErrorHandler, succHandler SuccessHandler)
}

// SpawnDaemon spawns the daemon process. args will be serialised and passed as command line arguments.
// env should contain entries of the form "key=value"; these will be available as environment variables.
// processName is used to define environment variables and file names. If the daemon process calls errHandler,
// the error will be returned by this function in the parent process. Otherwise, this function returns nil
// as soon as the child process calls succHandler.
func SpawnDaemon(daemonObj Daemon, env []string, processName string) error {
	environment.Init(processName)
	envInfo, err := error_handling.GetDaemonEnv()
	if err != nil {
		panic("Cannot setup daemon environment.\nError: " + err.Error())
	}
	defer error_handling.CleanupDeamonEnv(envInfo)

	env = append(envInfo.Env, env...)
	ctx := daemon.Context{
		Args:        argsToFullList(daemonObj),
		Env:         env,
		LogFileName: environment.LogFileName,
		LogFilePerm: 0755,
	}
	child, err := ctx.Reborn()
	if err != nil {
		return fmt.Errorf("cannot start the daemon process: %w", err)
	}

	if child == nil {
		// child code - run actual process
		childProcessPostSpawn(daemonObj, envInfo)
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

func childProcessPostSpawn(daemonObj Daemon, envInfo error_handling.EnvInfo) {
	sender, err := error_handling.NewSubprocessErrorSender(envInfo.NamedPipeName)
	if err != nil {
		panic("Unable to setup daemon error sender\nError: " + err.Error())
	}
	defer sender.Close()
	// run actual process code
	daemonObj.DaemonMain(sender, sender)
}
