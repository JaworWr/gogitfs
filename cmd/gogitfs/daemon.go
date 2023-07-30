package main

import (
	"fmt"
	"github.com/hanwen/go-fuse/v2/fs"
	"gogitfs/pkg/daemon"
	"gogitfs/pkg/daemon/environment"
	"gogitfs/pkg/error_handler"
	"gogitfs/pkg/gitfs"
	"gogitfs/pkg/logging"
	"math"
	"os/user"
	"strconv"
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
	fsOpts, err := getFuseOpts(opts)
	if err != nil {
		errHandler.HandleError(err)
	}
	fsOpts.AttrTimeout = &posTime
	fsOpts.EntryTimeout = &posTime
	fsOpts.NegativeTimeout = &negTime

	server, err := fs.Mount(opts.mountDir, root, fsOpts)
	if err != nil {
		errHandler.HandleError(err)
	}
	succHandler.HandleSuccess()
	server.Wait()
	logging.InfoLog.Printf("Exiting")
}

var _ daemon.ProcessInfo = (*gogitfsDaemon)(nil)

func getFuseOpts(o *daemonOptions) (*fs.Options, error) {
	opts := &fs.Options{}
	// get current UID and GID if not specified
	opts.UID = uint32(o.uid)
	opts.GID = uint32(o.gid)
	if opts.UID == math.MaxUint32 || opts.GID == math.MaxUint32 {
		currentUser, err := user.Current()
		if err != nil {
			return nil, err
		}
		if opts.UID == math.MaxUint32 {
			uid, err := strconv.ParseUint(currentUser.Uid, 10, 32)
			if err != nil {
				panic("Cannot parse UID.\nError: " + err.Error())
			}
			opts.UID = uint32(uid)
		}
		if opts.GID == math.MaxUint32 {
			gid, err := strconv.ParseUint(currentUser.Gid, 10, 32)
			if err != nil {
				panic("Cannot parse UID.\nError: " + err.Error())
			}
			opts.GID = uint32(gid)
		}
	}

	if o.fuseLog {
		fuseLogPath := o.fuseLogPath
		if fuseLogPath == "" {
			fuseLogPath = fmt.Sprintf("%s-%d-fuse.log", environment.DaemonName, environment.DaemonParentPid)
		}
		logger, err := logging.MakeFileLogger(fuseLogPath)
		if err != nil {
			return nil, err
		}
		opts.Logger = logger
		opts.Debug = o.fuseDebugLog
	}

	return opts, nil
}
