package main

import (
	"gogitfs/pkg/daemon"
	"log"
)

func main() {
	daemonInfo := &gogitfsDaemon{}
	err := daemon.SpawnDaemon(daemonInfo, "gogitfs")
	if err != nil {
		log.Fatalf("An error occurred - cannot start the filesystem daemon.\n%v", err.Error())
	}

}
