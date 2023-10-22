// Package main defines the CLI and the program's entry point.
package main

import (
	"flag"
	"fmt"
	"gogitfs/pkg/daemon"
	"os"
)

func main() {
	var err error

	daemonObj := &gogitfsDaemon{}
	err = daemon.ParseFlags(daemonObj, func() {

	})
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		flag.Usage()
		os.Exit(2)
	}

	err = daemon.SpawnDaemon(daemonObj, nil, "gogitfs")
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "cannot start the filesystem daemon\n%v\n", err.Error())
		os.Exit(1)
	}
}
