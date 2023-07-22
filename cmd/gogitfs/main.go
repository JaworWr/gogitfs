package main

import (
	"gogitfs/pkg/daemon"
	"log"
)

func main() {
	var err error

	daemonInfo := &gogitfsDaemon{}
	da := daemonInfo.DaemonArgs()
	err = parseArgs(da)
	if err != nil {
		log.Fatalln(err.Error())
	}

	err = daemon.SpawnDaemon(daemonInfo, "gogitfs")
	if err != nil {
		log.Fatalf("Cannot start the filesystem daemon.\n%v", err.Error())
	}
}
