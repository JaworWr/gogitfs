package main

import (
	"flag"
	"github.com/sevlyar/go-daemon"
)

var shouldShowHelp bool

func setupHelp() {
	if daemon.WasReborn() {
		return
	}
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
