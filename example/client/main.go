package main

import (
	"github.com/hackbeex/configcenter/client"
	"github.com/hackbeex/configcenter/util/log"
)

func main() {
	cl := client.New(&client.Config{
		ClientHost:    "127.0.0.1",
		ClientPort:    8888,
		ClientCluster: "default",
		ClientApp:     "test_app",
		ClientEnv:     "dev",
		DiscoverHost:  "127.0.0.1",
		DiscoverPort:  9310,
	})
	if err := cl.Register(); err != nil {
		log.Fatal(err)
	}

	go cl.WatchServerExit()

	configs, err := cl.GetAllConfig()
	if err != nil {
		log.Fatal(err)
	}
	log.Info("configs: ", configs)

	val, ok := cl.GetConfig("test", "123")
	log.Infof("is_exist: %t, value: %s", ok, val)

	cl.ListenConfig("test", func(param *client.CallbackParam) {
		log.Infof("listen: key:%s, val:%s, type: %s", param.Key, param.NewVal, param.OpType)
	})

	//global listen
	cl.ListenConfig("", func(param *client.CallbackParam) {
		log.Infof("listen: key:%s, val:%s, type: %s", param.Key, param.NewVal, param.OpType)
	})

	//ctrl+C to exit
	select {}
}
