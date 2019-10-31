package discover

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hackbeex/configcenter/discover/handler"
	"github.com/hackbeex/configcenter/discover/store"
	"github.com/hackbeex/configcenter/local"
	"github.com/hackbeex/configcenter/util/log"
)

var table *Table

func Run() {
	table = initTable()
	runServer()
}

func runServer() {
	r := gin.Default()

	r.POST("/api/v1/discover/server/register", handler.ServerRegister)
	r.POST("/api/v1/discover/server/heartbeat", handler.ServerHeartbeat)
	r.POST("/api/v1/discover/server/fetch", handler.ServerFetch)
	r.POST("/api/v1/discover/client/register", handler.ClientRegister)
	r.POST("/api/v1/discover/client/heartbeat", handler.ClientHeartbeat)
	r.POST("/api/v1/discover/client/fetch", handler.ClientFetch)
	r.POST("/api/v1/discover/portal/register", handler.PortalRegister)
	r.POST("/api/v1/discover/portal/heartbeat", handler.PortalHeartbeat)

	conf := local.Conf.Discover
	addr := fmt.Sprintf("%s:%d", conf.ListenHost, conf.ListenPort)
	log.Infof("discover server run at: %s", addr)

	if err := r.Run(addr); err != nil {
		log.Panic(err)
	}
}

func GetTable() *Table {
	return table
}

func GetStore() *store.Store {
	return table.GetStore()
}
