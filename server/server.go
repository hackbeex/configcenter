package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hackbeex/configcenter/local"
	"github.com/hackbeex/configcenter/util/log"
)

func Run() {
	runServer()
}

func runServer() {
	r := gin.Default()

	r.POST("/api/v1/app/list")
	r.POST("/api/v1/app/detail")
	r.POST("/api/v1/app/create")
	//r.POST("/api/v1/cluster/create", )
	r.POST("/api/v1/namespace/create")
	r.POST("/api/vi/config/detail")
	r.POST("/api/vi/config/create")
	r.POST("/api/vi/config/update")
	r.POST("/api/vi/config/delete")
	r.POST("/api/vi/config/release")
	r.POST("/api/vi/config/rollback")
	r.POST("/api/vi/config/sync")
	r.POST("/api/vi/config/history")

	conf := local.Conf.Server
	addr := fmt.Sprintf("%s:%d", conf.ListenHost, conf.ListenPort)
	log.Infof("config server run at: %s", addr)

	if err := r.Run(addr); err != nil {
		log.Panic(err)
	}
}
