package main

import (
	"github.com/hanwen/go-fuse/v2/fs"
	"gogitfs/pkg/daemon"
	"gogitfs/pkg/error_handler"
	"gogitfs/pkg/gitfs"
	"gogitfs/pkg/logging"
	"gogitfs/pkg/mountpoint"
	"math"
	"os/user"
	"strconv"
	"time"
)

func (d *gogitfsDaemon) DaemonMain(
	errHandler error_handler.ErrorHandler,
	succHandler daemon.SuccessHandler,
) {
	logging.Init(d.logLevel)
	errHandler = error_handler.MakeLoggingHandler(errHandler, logging.Error)
	logging.InfoLog.Printf("Log level: %v\n", d.logLevel.String())
	logging.InfoLog.Printf("Repository path: %v\n", d.repoDir)
	root, err := gitfs.NewRootNode(d.repoDir)
	if err != nil {
		errHandler.HandleError(err)
	}

	mountDir, err := mountpoint.ValidateMountpoint(d.mountDir, d.allowNonEmpty)
	if err != nil {
		errHandler.HandleError(err)
	}
	logging.InfoLog.Printf("Mounting in %v\n", mountDir)

	posTime := 6 * time.Hour
	negTime := 15 * time.Second
	fsOpts, err := getFuseOpts(d)
	if err != nil {
		errHandler.HandleError(err)
	}
	fsOpts.AttrTimeout = &posTime
	fsOpts.EntryTimeout = &posTime
	fsOpts.NegativeTimeout = &negTime
	fsOpts.Logger = logging.ErrorLog

	server, err := fs.Mount(mountDir, root, fsOpts)
	if err != nil {
		errHandler.HandleError(err)
	}
	succHandler.HandleSuccess()
	server.Wait()
	logging.InfoLog.Printf("Exiting")
}

var _ daemon.Daemon = (*gogitfsDaemon)(nil)

func getFuseOpts(d *gogitfsDaemon) (*fs.Options, error) {
	opts := &fs.Options{}
	// get current UID and GID if not specified
	opts.UID = uint32(d.uid)
	opts.GID = uint32(d.gid)
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

	opts.Debug = d.fuseDebug
	return opts, nil
}
