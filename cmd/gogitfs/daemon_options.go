package main

import (
	"flag"
	"gogitfs/pkg/daemon"
	"gogitfs/pkg/logging"
)

const (
	logLevelFlag     = "log-level"
	fuseLogFlag      = "fuse-log"
	fuseLogPathFlag  = "fuse-log-path"
	fuseDebugLogFlag = "fuse-log-debug"
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
	flag.Var(&o.logLevel, logLevelFlag, "log level, can be given as upper-case string or an integer")

	flag.BoolVar(&o.fuseLog, fuseLogFlag, false, "enable FUSE log")
	flag.StringVar(&o.fuseLogPath, fuseLogPathFlag, "", "FUSE log file name")
	flag.BoolVar(&o.fuseDebugLog, fuseDebugLogFlag, false, "enable FUSE debug log")
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
		daemon.SerializeStringFlag(logLevelFlag, o.logLevel.String()),
		daemon.SerializeBoolFlag(fuseLogFlag, o.fuseLog),
		daemon.SerializeStringFlag(fuseLogPathFlag, o.fuseLogPath),
		daemon.SerializeBoolFlag(fuseDebugLogFlag, o.fuseDebugLog),
		o.repoDir,
		o.mountDir,
	}
}

var _ daemon.DaemonArgs = (*daemonOptions)(nil)
