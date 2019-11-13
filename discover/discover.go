package discover

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hackbeex/configcenter/discover/handler"
	"github.com/hackbeex/configcenter/discover/meta"
	"github.com/hackbeex/configcenter/local"
	"github.com/hackbeex/configcenter/util/log"
)

func Run() {
	meta.InitTable()
	runServer()
}

func runServer() {
	r := gin.Default()

	r.POST("/api/v1/discover/server/register", handler.ServerRegister)
	r.POST("/api/v1/discover/server/heartbeat", handler.ServerHeartbeat)
	r.POST("/api/v1/discover/server/fetch", handler.ServerFetch)
	r.POST("/api/v1/discover/client/register", handler.ClientRegister) //no use
	r.POST("/api/v1/discover/client/fetch", handler.ClientFetch)
	r.POST("/api/v1/discover/portal/register", handler.PortalRegister) //no use

	conf := local.Conf.Discover
	addr := fmt.Sprintf("%s:%d", conf.ListenHost, conf.ListenPort)
	log.Infof("discover server run at: %s", addr)

	if err := r.Run(addr); err != nil {
		log.Panic(err)
	}
}
