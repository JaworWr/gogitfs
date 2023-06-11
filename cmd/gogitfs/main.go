package main

import (
	"gogitfs/pkg/daemon"
	"gogitfs/pkg/error_handler"
)

func main() {
	daemonInfo := &gogitfsDaemon{}
	err := daemon.SpawnDaemon(daemonInfo, "gogitfs")
	if err != nil {
		error_handler.FatalHandler.HandleError(err)
	}
}
