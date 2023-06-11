package daemon

import (
	"github.com/sevlyar/go-daemon"
	"gogitfs/pkg/daemon/internal/error_handling"
	"gogitfs/pkg/error_handler"
	"log"
	"os"
)

type ProcessInfo interface {
	DaemonArgs(args []string) []string
	DaemonEnv(env []string) []string
	DaemonProcess(errHandler error_handler.ErrorHandler, succHandler SuccessHandler)
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
		childProcessPostSpawn(info)
		return nil
	}
	// parent code - handle errors from child
	err = parentProcessPostSpawn(envInfo)
	return err
}

func parentProcessPostSpawn(envInfo *error_handling.EnvInfo) error {
	receiver, err := error_handling.NewSubprocessErrorReceiver(envInfo.NamedPipeName)
	if err != nil {
		log.Panicf("Unable to setup daemon error receiver\n%v", err.Error())
	}
	defer receiver.Close()
	return receiver.Receive()
}

func childProcessPostSpawn(info ProcessInfo) {
	sender, err := error_handling.NewSubprocessErrorSender()
	if err != nil {
		log.Panicf("Unable to setup daemon error sender\n%v", err.Error())
	}
	defer (func() {
		err := sender.Close()
		if err != nil {
			log.Printf("Sender close error\n%v", err.Error())
		}
	})()
	info.DaemonProcess(sender, sender)
}
