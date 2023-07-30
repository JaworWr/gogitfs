package main

import (
	"flag"
	"gogitfs/pkg/daemon"
	"gogitfs/pkg/logging"
	"math"
)

const (
	logLevelFlag      = "log-level"
	fuseDebugFlag     = "fuse-debug"
	allowNonEmptyFlag = "allow-nonempty"
	uidFlag           = "uid"
	gidFlag           = "gid"
	allowOtherFlag    = "allow-other"
)

type daemonOptions struct {
	repoDir  string
	mountDir string

	logLevel  logging.LogLevelFlag
	fuseDebug bool

	allowNonEmpty bool

	uid        uint
	gid        uint
	allowOther bool
}

func (o *daemonOptions) Setup() {
	o.logLevel = logging.Info
	flag.Var(&o.logLevel, logLevelFlag, "log level, can be given as upper-case string or an integer")
	flag.BoolVar(&o.fuseDebug, fuseDebugFlag, false, "show FUSE debug info in logs")

	flag.BoolVar(&o.allowNonEmpty, allowNonEmptyFlag, false, "allow mounting in a non-empty directory")

	flag.UintVar(&o.uid, uidFlag, math.MaxUint32, "UID (user ID) to mount as")
	flag.UintVar(&o.gid, gidFlag, math.MaxUint32, "GID (group ID) to mount as")
	flag.BoolVar(&o.allowOther, allowOtherFlag, false, "mount FUSE filesystem with 'allow_other'")
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
		daemon.SerializeBoolFlag(fuseDebugFlag, o.fuseDebug),
		daemon.SerializeBoolFlag(allowNonEmptyFlag, o.allowNonEmpty),
		daemon.SerializeUintFlag(uidFlag, uint64(o.uid)),
		daemon.SerializeUintFlag(gidFlag, uint64(o.gid)),
		daemon.SerializeBoolFlag(allowOtherFlag, o.allowOther),
		o.repoDir,
		o.mountDir,
	}
}

var _ daemon.DaemonArgs = (*daemonOptions)(nil)
