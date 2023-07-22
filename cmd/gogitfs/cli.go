package main

import "flag"

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
