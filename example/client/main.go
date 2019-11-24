package client

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

	configs, err := cl.GetAllConfig()
	if err != nil {
		log.Fatal(err)
	}
	log.Info("configs: ", configs)

	val, ok := cl.GetConfig("test", "123")
	log.Infof("is_exsit: %T, value: %s", ok, val)

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
