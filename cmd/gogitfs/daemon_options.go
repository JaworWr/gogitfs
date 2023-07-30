package main

import (
	"flag"
	"fmt"
	"gogitfs/pkg/daemon"
	"gogitfs/pkg/logging"
)

type daemonOptions struct {
	repoDir  string
	mountDir string

	logLevel logging.LogLevelFlag

	fuseLog      bool
	fuseLogPath  string
	fuseDebugLog bool
}

func (o *daemonOptions) Setup() {
	o.logLevel = logging.Info
	flag.Var(&o.logLevel, "log-level", "log level, can be given as upper-case string or an integer")

	flag.BoolVar(&o.fuseLog, "fuse-log", false, "enable FUSE log")
	flag.StringVar(&o.fuseLogPath, "fuse-log-path", "", "FUSE log file name")
	flag.BoolVar(&o.fuseDebugLog, "fuse-log-debug", false, "enable FUSE debug log")
}

func (o *daemonOptions) PositionalArgs() []daemon.PositionalArg {
	return []daemon.PositionalArg{
		{Name: "repo-dir", Usage: "directory of the repository"},
		{Name: "mount-dir", Usage: "where to mount the repository"},
	}
}

func (o *daemonOptions) HandlePositionalArgs(args []string) error {
	if len(args) < 2 {
		return &daemon.NotEnoughArgsError{Expected: 2, Got: len(args)}
	} else if len(args) > 2 {
		return &daemon.TooManyArgsError{ExtraArgs: args[2:]}
	}
	o.repoDir = args[0]
	o.mountDir = args[1]
	return nil
}

func (o *daemonOptions) Serialize() []string {
	return []string{
		fmt.Sprintf("--loglevel=%v", o.logLevel.String()),
		o.repoDir,
		o.mountDir,
	}
}

var _ daemon.DaemonArgs = (*daemonOptions)(nil)
