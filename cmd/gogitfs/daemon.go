package main

import (
	"github.com/hanwen/go-fuse/v2/fs"
	"gogitfs/pkg/daemon"
	"gogitfs/pkg/error_handler"
	"gogitfs/pkg/gitfs"
	"gogitfs/pkg/logging"
	"time"
)

type gogitfsDaemon struct{}

func (g *gogitfsDaemon) DaemonArgs(args []string) []string {
	return args
}

func (g *gogitfsDaemon) DaemonEnv(_ []string) []string {
	return nil
}

func (g *gogitfsDaemon) DaemonProcess(errHandler error_handler.ErrorHandler, succHandler daemon.SuccessHandler) {
	opts, err := parseDaemonOpts()
	if err != nil {
		errHandler.HandleError(err)
	}
	logging.Init(opts.logLevel)
	errHandler = error_handler.MakeLoggingHandler(errHandler)
	logging.InfoLog.Printf("Log level: %v\n", opts.logLevel)
	logging.InfoLog.Printf("Repository path: %v\n", opts.repoDir)
	root, err := gitfs.NewRootNode(opts.repoDir)
	if err != nil {
		errHandler.HandleError(err)
	}
	logging.InfoLog.Printf("Mounting in %v\n", opts.mountDir)
	h := 6 * time.Hour
	fsOpts := fs.Options{
		AttrTimeout:  &h,
		EntryTimeout: &h,
	}
	server, err := fs.Mount(opts.mountDir, root, &fsOpts)
	if err != nil {
		errHandler.HandleError(err)
	}
	succHandler.HandleSuccess()
	server.Wait()
	logging.InfoLog.Printf("Exiting")
}

var _ daemon.ProcessInfo = (*gogitfsDaemon)(nil)
