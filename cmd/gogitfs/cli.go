package main

import (
	"flag"
	goDaemon "github.com/sevlyar/go-daemon"
	"gogitfs/pkg/daemon"
)

var shouldShowHelp bool

func setupHelp() {
	flag.BoolVar(&shouldShowHelp, "help", false, "show help and exit")
	flag.BoolVar(&shouldShowHelp, "h", false, "shorthand for --help")
}

func showHelp() {
	if !shouldShowHelp {
		return
	}
	// for now - just show usage
	flag.Usage()
}

func parseArgs(da daemon.DaemonArgs) error {
	if goDaemon.WasReborn() {
		return nil
	}
	setupHelp()
	daemon.InitArgs(da)
	flag.Parse()
	showHelp()
	err := da.HandlePositionalArgs(flag.Args())
	return err
}
