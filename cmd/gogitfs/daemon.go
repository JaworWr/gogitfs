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

func (g *gogitfsDaemon) DaemonArgs() daemon.DaemonArgs {
	return &daemonOptions{}
}

func (g *gogitfsDaemon) DaemonEnv() []string {
	return nil
}

func (g *gogitfsDaemon) DaemonProcess(
	args daemon.DaemonArgs,
	errHandler error_handler.ErrorHandler,
	succHandler daemon.SuccessHandler,
) {
	opts := args.(*daemonOptions)
	logging.Init(opts.logLevel)
	errHandler = error_handler.MakeLoggingHandler(errHandler, logging.Error)
	logging.InfoLog.Printf("Log level: %v\n", opts.logLevel.String())
	logging.InfoLog.Printf("Repository path: %v\n", opts.repoDir)
	root, err := gitfs.NewRootNode(opts.repoDir)
	if err != nil {
		errHandler.HandleError(err)
	}
	logging.InfoLog.Printf("Mounting in %v\n", opts.mountDir)
	posTime := 6 * time.Hour
	negTime := 15 * time.Second
	fsOpts := fs.Options{
		AttrTimeout:     &posTime,
		EntryTimeout:    &posTime,
		NegativeTimeout: &negTime,
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
