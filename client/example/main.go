package main

import (
	"github.com/hackbeex/configcenter/client"
	"github.com/hackbeex/configcenter/util/log"
)

func main() {
	cl := &client.Client{
		Host:    "127.0.0.1",
		Port:    8888,
		Cluster: "default",
		Env:     "dev",
	}
	if err := cl.Register(); err != nil {
		log.Fatal(err)
	}

	go cl.WatchServerExit()
}
