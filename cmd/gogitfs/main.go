package main

import (
	"flag"
	"fmt"
	"gogitfs/pkg/daemon"
	"os"
)

func main() {
	var err error

	daemonInfo := &gogitfsDaemon{}
	da := daemonInfo.DaemonArgs()
	err = parseArgs(da)
	if err != nil {
		fmt.Println(err.Error())
		flag.Usage()
		os.Exit(2)
	}

	err = daemon.SpawnDaemon(da, daemonInfo, "gogitfs")
	if err != nil {
		fmt.Printf("cannot start the filesystem daemon\n%v", err.Error())
		os.Exit(1)
	}
}
