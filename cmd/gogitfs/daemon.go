package main

import (
	"fmt"
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
	errHandler = error_handler.MakeLoggingHandler(errHandler)
	if len(os.Args) < 3 {
		err := fmt.Errorf("not enough arguments. Usage: gogitfs <repo-path> <mount-path>")
		errHandler.HandleError(err)
	}
	repoDir := os.Args[1]
	mountDir := os.Args[2]
	root, err := gitfs.NewRootNode(repoDir)
	if err != nil {
		errHandler.HandleError(err)
	}
	log.Printf("Mounting in %v\n", mountDir)
	server, err := fs.Mount(mountDir, root, &fs.Options{})
	if err != nil {
		errHandler.HandleError(err)
	}
	succHandler.HandleSuccess()
	server.Wait()
	log.Printf("Exiting")
}

var _ daemon.ProcessInfo = (*gogitfsDaemon)(nil)
