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

type gogitfsDaemon struct {
	repoDir  string
	mountDir string

	logLevel  logging.LogLevelFlag
	fuseDebug bool

	allowNonEmpty bool

	uid        uint
	gid        uint
	allowOther bool
}

func (d *gogitfsDaemon) Setup() {
	d.logLevel = logging.Info
	flag.Var(&d.logLevel, logLevelFlag, "log level, can be given as upper-case string or an integer")
	flag.BoolVar(&d.fuseDebug, fuseDebugFlag, false, "show FUSE debug info in logs")

	flag.BoolVar(&d.allowNonEmpty, allowNonEmptyFlag, false, "allow mounting in a non-empty directory")

	flag.UintVar(&d.uid, uidFlag, math.MaxUint32, "UID (user ID) to mount as")
	flag.UintVar(&d.gid, gidFlag, math.MaxUint32, "GID (group ID) to mount as")
	flag.BoolVar(&d.allowOther, allowOtherFlag, false, "mount FUSE filesystem with 'allow_other'")
}

func (d *gogitfsDaemon) PositionalArgs() []daemon.PositionalArg {
	return []daemon.PositionalArg{
		{Name: "repo-dir", Usage: "directory of the repository"},
		{Name: "mount-dir", Usage: "where to mount the repository"},
	}
}

func (d *gogitfsDaemon) HandlePositionalArgs(args []string) error {
	if len(args) < 2 {
		return &daemon.NotEnoughArgsError{Expected: 2, Got: len(args)}
	} else if len(args) > 2 {
		return &daemon.TooManyArgsError{ExtraArgs: args[2:]}
	}
	d.repoDir = args[0]
	d.mountDir = args[1]
	return nil
}

func (d *gogitfsDaemon) Serialize() []string {
	return []string{
		daemon.SerializeStringFlag(logLevelFlag, d.logLevel.String()),
		daemon.SerializeBoolFlag(fuseDebugFlag, d.fuseDebug),
		daemon.SerializeBoolFlag(allowNonEmptyFlag, d.allowNonEmpty),
		daemon.SerializeUintFlag(uidFlag, uint64(d.uid)),
		daemon.SerializeUintFlag(gidFlag, uint64(d.gid)),
		daemon.SerializeBoolFlag(allowOtherFlag, d.allowOther),
		d.repoDir,
		d.mountDir,
	}
}

var _ daemon.CliArgs = (*gogitfsDaemon)(nil)
