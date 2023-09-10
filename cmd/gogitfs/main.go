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
		fmt.Println(err.Error())
		flag.Usage()
		os.Exit(2)
	}

	err = daemon.SpawnDaemon(daemonObj, nil, "gogitfs")
	if err != nil {
		fmt.Printf("cannot start the filesystem daemon\n%v\n", err.Error())
		os.Exit(1)
	}
}
