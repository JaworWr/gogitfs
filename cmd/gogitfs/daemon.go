package main

import (
	"flag"
	"fmt"
	"github.com/hanwen/go-fuse/v2/fs"
	"gogitfs/pkg/daemon"
	"gogitfs/pkg/error_handler"
	"gogitfs/pkg/gitfs"
	"gogitfs/pkg/logging"
	"log"
	"time"
)

type gogitfsDaemon struct{}

func (g *gogitfsDaemon) DaemonArgs(args []string) []string {
	return args
}

func (g *gogitfsDaemon) DaemonEnv(_ []string) []string {
	return nil
}

type options struct {
	repoDir  string
	mountDir string
	logLevel logging.LogLevelFlag
}

func (o *options) parse(errHandler error_handler.ErrorHandler) {
	flag.IntVar((*int)(&o.logLevel), "loglevel", int(logging.Info), "log level")
	flag.Parse()
	if flag.NArg() < 2 {
		err := fmt.Errorf("not enough arguments. Usage: gogitfs <repo-path> <mount-path>")
		errHandler.HandleError(err)
	}
	o.repoDir = flag.Arg(0)
	o.mountDir = flag.Arg(1)
}

func (g *gogitfsDaemon) DaemonProcess(errHandler error_handler.ErrorHandler, succHandler daemon.SuccessHandler) {
	errHandler = error_handler.MakeLoggingHandler(errHandler)
	opts := options{}
	opts.parse(errHandler)
	log.Printf("Log level: %v\n", opts.logLevel)
	log.Printf("Repository path: %v\n", opts.repoDir)
	root, err := gitfs.NewRootNode(opts.repoDir)
	if err != nil {
		errHandler.HandleError(err)
	}
	log.Printf("Mounting in %v\n", opts.mountDir)
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
	log.Printf("Exiting")
}

var _ daemon.ProcessInfo = (*gogitfsDaemon)(nil)
