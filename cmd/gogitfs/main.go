package main

import (
	"flag"
	"gogitfs/pkg/daemon"
	"log"
)

func main() {
	var err error

	setupHelp()
	daemonInfo := &gogitfsDaemon{}
	da := daemonInfo.DaemonArgs()
	daemon.InitArgs(da)
	flag.Parse()
	err = da.HandlePositionalArgs(flag.Args())
	if err != nil {
		log.Fatalln(err.Error())
	}

	err = daemon.SpawnDaemon(daemonInfo, "gogitfs")
	if err != nil {
		log.Fatalf("Cannot start the filesystem daemon.\n%v", err.Error())
	}
}
