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
	flag.IntVar((*int)(&opts.logLevel), "loglevel", int(logging.Info), "log level")
	flag.Parse()
	if flag.NArg() < 2 {
		err = fmt.Errorf("not enough arguments")
		return
	}
	opts.repoDir = flag.Arg(0)
	opts.mountDir = flag.Arg(1)
	return
}
