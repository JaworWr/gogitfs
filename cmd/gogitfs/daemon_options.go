package main

import (
	"flag"
	"gogitfs/pkg/daemon"
	"gogitfs/pkg/logging"
)

// CLI flag names
const (
	logLevelFlag      = "log-level"
	fuseDebugFlag     = "fuse-debug"
	allowNonEmptyFlag = "allow-nonempty"
	uidFlag           = "uid"
	gidFlag           = "gid"
	allowOtherFlag    = "allow-other"
)

// gogitfsDaemon describes a daemon process handling the mounted repository
// When started, it runs a FUSE server. The process exits as soon as the server stops,
// which happens when the directory is unmounted.
type gogitfsDaemon struct {
	repoDir  string
	mountDir string

	logLevel  logging.LogLevelFlag
	fuseDebug bool

	allowNonEmpty bool

	uid        int64
	gid        int64
	allowOther bool
}

func (d *gogitfsDaemon) Setup() {
	d.logLevel = logging.Info
	flag.Var(&d.logLevel, logLevelFlag, "log level, can be given as upper-case string or an integer")
	flag.BoolVar(&d.fuseDebug, fuseDebugFlag, false, "show FUSE debug info in logs")

	flag.BoolVar(&d.allowNonEmpty, allowNonEmptyFlag, false, "allow mounting in a non-empty directory")

	flag.Int64Var(&d.uid, uidFlag, -1, "UID (user ID) to mount as; pass -1 to use current user's ID")
	flag.Int64Var(&d.gid, gidFlag, -1, "GID (group ID) to mount as; pass -1 to use current user group's ID")
	flag.BoolVar(&d.allowOther, allowOtherFlag, false, "mount FUSE filesystem with 'allow_other'")
}

func (d *gogitfsDaemon) PositionalArgs() []daemon.PositionalArg {
	return []daemon.PositionalArg{
		{Name: "repo-dir", Usage: "path to the repository"},
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
		daemon.SerializeIntFlag(uidFlag, d.uid),
		daemon.SerializeIntFlag(gidFlag, d.gid),
		daemon.SerializeBoolFlag(allowOtherFlag, d.allowOther),
		d.repoDir,
		d.mountDir,
	}
}

var _ daemon.SerializableCliArgs = (*gogitfsDaemon)(nil)
