package main

import (
	"flag"
	"fmt"
	"gogitfs/pkg/logging"
)

type daemonOptions struct {
	repoDir  string
	mountDir string
	logLevel logging.LogLevelFlag
}

func parseDaemonOpts() (opts daemonOptions, err error) {
	opts.logLevel = logging.Info
	flag.Var(&opts.logLevel, "loglevel", "log level")
	flag.Parse()
	if flag.NArg() < 2 {
		err = fmt.Errorf("not enough arguments")
		return
	}
	opts.repoDir = flag.Arg(0)
	opts.mountDir = flag.Arg(1)
	return
}
