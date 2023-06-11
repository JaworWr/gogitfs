package main

import (
	"github.com/hanwen/go-fuse/v2/fs"
	"gogitfs/pkg/daemon"
	"gogitfs/pkg/error_handler"
	"gogitfs/pkg/gitfs"
	"log"
	"os"
)

type gogitfsDaemon struct{}

func (g *gogitfsDaemon) DaemonArgs(args []string) []string {
	return args
}

func (g *gogitfsDaemon) DaemonEnv(_ []string) []string {
	return nil
}

func (g *gogitfsDaemon) DaemonProcess(errHandler error_handler.ErrorHandler, succHandler daemon.SuccessHandler) {
	mountDir := os.Args[1]
	root := &gitfs.RootNode{}
	log.Printf("Mounting in %v\n", mountDir)
	server, err := fs.Mount(mountDir, root, &fs.Options{})
	if err != nil {
		errHandler.HandleError(err)
	}
	succHandler.HandleSuccess()
	server.Wait()
}

var _ daemon.ProcessInfo = (*gogitfsDaemon)(nil)
