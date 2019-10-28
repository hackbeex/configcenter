package discover

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hackbeex/configcenter/local"
	"github.com/hackbeex/configcenter/util/log"
	"os"
)

var table *Table

func Run() {
	table = initTable()
	go runServer()
}

func runServer() {
	r := gin.Default()

	//route
	r.POST("/api/v1/server/register")
	r.POST("/api/v1/server/heartbeat")
	r.POST("/api/v1/server/fetch")
	r.POST("/api/v1/client/register")
	r.POST("/api/v1/client/heartbeat")
	r.POST("/api/v1/client/fetch")
	r.POST("/api/v1/portal/register")
	r.POST("/api/v1/portal/heartbeat")
	r.POST("/api/v1/config/fetch")

	conf := local.Conf.Discover
	addr := fmt.Sprintf("%s:%d", conf.ListenHost, conf.ListenPort)
	log.Infof("config server run at: %s", addr)

	if err := r.Run(addr); err != nil {
		log.Error(err)
		os.Exit(-1)
	}
}

func GetTable() *Table {
	return table
}
