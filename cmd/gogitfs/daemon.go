package main

import (
	"fmt"
	"gogitfs/pkg/daemon"
	"gogitfs/pkg/error_handler"
	"time"
)

type gogitfsDaemon struct{}

func (g gogitfsDaemon) DaemonArgs(args []string) []string {
	return nil
}

func (g gogitfsDaemon) DaemonEnv(env []string) []string {
	return nil
}

func (g gogitfsDaemon) DaemonProcess(errHandler error_handler.ErrorHandler, succHandler daemon.SuccessHandler) {
	fmt.Println("Hello from daemon")
	time.Sleep(3 * time.Second)
	fmt.Println("Bye")
	succHandler.HandleSuccess()
}

var _ daemon.ProcessInfo = (*gogitfsDaemon)(nil)
