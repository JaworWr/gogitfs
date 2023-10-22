package main

import (
	"fmt"
	"github.com/hanwen/go-fuse/v2/fs"
	"gogitfs/pkg/daemon"
	"gogitfs/pkg/error_handler"
	"gogitfs/pkg/gitfs"
	"gogitfs/pkg/logging"
	"gogitfs/pkg/mountpoint"
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
		err = fmt.Errorf("cannot create root node: %w", err)
		errHandler.HandleError(err)
	}

	mountDir, err := mountpoint.ValidateMountpoint(d.mountDir, d.allowNonEmpty)
	if err != nil {
		err = fmt.Errorf("invalid mountpoint: %w", err)
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
		err = fmt.Errorf("cannot start FUSE server: %w", err)
		errHandler.HandleError(err)
	}
	succHandler.HandleSuccess()
	server.Wait()
	logging.InfoLog.Printf("Exiting")
}

var _ daemon.Daemon = (*gogitfsDaemon)(nil)

// getFuseOpts sets FUSE options based on the daemon's CLI arguments.
func getFuseOpts(d *gogitfsDaemon) (*fs.Options, error) {
	opts := &fs.Options{}
	// get current UID and GID if not specified
	if d.uid == -1 || d.gid == -1 {
		currentUser, err := user.Current()
		if err != nil {
			return nil, fmt.Errorf("cannot get current user: %w", err)
		}
		if d.uid == -1 {
			uid, err := strconv.ParseUint(currentUser.Uid, 10, 32)
			if err != nil {
				panic("Cannot parse UID.\nError: " + err.Error())
			}
			opts.UID = uint32(uid)
		}
		if d.gid == -1 {
			gid, err := strconv.ParseUint(currentUser.Gid, 10, 32)
			if err != nil {
				panic("Cannot parse UID.\nError: " + err.Error())
			}
			opts.GID = uint32(gid)
		}
	} else {
		opts.UID = uint32(d.uid)
		opts.GID = uint32(d.gid)
	}

	opts.Debug = d.fuseDebug
	return opts, nil
}
