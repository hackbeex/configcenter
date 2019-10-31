package main

import (
	"github.com/hackbeex/configcenter/discover"
	"github.com/hackbeex/configcenter/server"
)

func main() {
	//todo: set run single or together mode
	go discover.Run()

	server.Run()
}
