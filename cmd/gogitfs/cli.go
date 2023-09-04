package main

import (
	"flag"
	goDaemon "github.com/sevlyar/go-daemon"
	"gogitfs/pkg/daemon"
	"gogitfs/pkg/daemon/environment"
	"os"
)

var shouldShowHelp bool

func setupHelp() {
	flag.BoolVar(&shouldShowHelp, "help", false, "show help and exit")
	flag.BoolVar(&shouldShowHelp, "h", false, "shorthand for --help")
}

func parseArgs(da daemon.DaemonArgs) error {
	daemon.SetupFlags(da)
	if !goDaemon.WasReborn() {
		setupHelp()
		environment.SetupFlags()
	}
	flag.Parse()

	if shouldShowHelp {
		flag.Usage()
		os.Exit(0)
	}

	err := da.HandlePositionalArgs(flag.Args())
	return err
}
