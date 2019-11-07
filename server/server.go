package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hackbeex/configcenter/local"
	"github.com/hackbeex/configcenter/server/handler"
	"github.com/hackbeex/configcenter/util"
	"github.com/hackbeex/configcenter/util/log"
	"net/http"
)

func Run(env string) {
	go registerServer(env)
	runServer()
}

func registerServer(env string) {
	type req struct {
		Id   string `json:"id"`
		Host string `json:"host"`
		Port int    `json:"port"`
		Env  string `json:"env"`
	}
	conf := local.Conf.Server
	id, err := util.GetUidFromHardwareAddress(conf.ListenPort)
	if err != nil {
		log.Fatal(err)
	}
	data, _ := json.Marshal(req{
		Id:   id,
		Host: conf.ListenHost,
		Port: conf.ListenPort,
		Env:  env,
	})
	url := "/api/v1/discover/server/register"
	_, err = http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		log.Fatal(err)
	}
	log.Info("config server register successful:", id)
}

func runServer() {
	r := gin.Default()

	r.POST("/api/v1/app/list", handler.GetAppList)
	r.POST("/api/v1/app/detail", handler.GetAppDetail)
	r.POST("/api/v1/app/create", handler.CreateApp)
	r.POST("/api/v1/cluster/create", handler.CreateCluster)
	r.POST("/api/v1/namespace/create", handler.CreateNamespace)
	r.POST("/api/v1/config/detail", handler.GetConfigDetail)
	r.POST("/api/v1/config/list", handler.GetConfigList)
	r.POST("/api/v1/config/create", handler.CreateConfig)
	r.POST("/api/v1/config/update", handler.UpdateConfig)
	r.POST("/api/v1/config/delete", handler.DeleteConfig)
	r.POST("/api/v1/config/history", handler.GetConfigHistory)
	r.POST("/api/v1/config/release", handler.ReleaseConfig)
	r.POST("/api/v1/config/release/history", handler.GetConfigReleaseHistory)
	r.POST("/api/v1/config/rollback", handler.RollbackConfig)
	r.POST("/api/v1/config/sync", handler.SyncConfig)
	r.POST("/api/v1/config/watch", handler.WatchConfig)
	r.POST("/api/v1/instance/list", handler.GetInstanceList)

	conf := local.Conf.Server
	addr := fmt.Sprintf("%s:%d", conf.ListenHost, conf.ListenPort)
	log.Infof("config server run at: %s", addr)

	if err := r.Run(addr); err != nil {
		log.Panic(err)
	}
}
