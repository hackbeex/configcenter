package main

import (
	"github.com/hackbeex/configcenter/discover"
	"github.com/hackbeex/configcenter/server"
	"os"
)

func main() {
	//todo: set run single or together mode
	go discover.Run()

	env := os.Getenv("ConfigCenterEnv")

	server.Run(env)
}
