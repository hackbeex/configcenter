package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/hackbeex/configcenter/discover"
	"github.com/hackbeex/configcenter/discover/com"
	"github.com/hackbeex/configcenter/discover/server"
	"github.com/hackbeex/configcenter/util/response"
)

func ServerRegister(c *gin.Context) {
	var req struct {
		Id   string
		Host string
		Port int
		Env  server.EnvType
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	svr := server.Server{
		Id:   req.Id,
		Host: req.Host,
		Port: req.Port,
		Env:  req.Env,
	}
	if err := svr.Register(discover.GetStore()); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c)
}

func ServerHeartbeat(c *gin.Context) {
	var req struct {
		Id server.IdKey
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	servers := discover.GetTable().Servers()
	if err := servers.UpdateStatus(req.Id, com.OnlineStatus); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c)
}

func ServerFetch(c *gin.Context) {
	servers := discover.GetTable().Servers()
	res, err := servers.FetchServerList()
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Data(c, map[string]interface{}{"list": res})
}
