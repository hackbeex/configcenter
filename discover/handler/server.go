package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/hackbeex/configcenter/discover/meta"
	"github.com/hackbeex/configcenter/discover/server"
	"github.com/hackbeex/configcenter/util/com"
	"github.com/hackbeex/configcenter/util/response"
)

func ServerRegister(c *gin.Context) {
	var req struct {
		Id   string      `json:"id"`
		Host string      `json:"host"`
		Port int         `json:"port"`
		Env  com.EnvType `json:"env"`
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
	if err := svr.Register(meta.GetStore()); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c)
}

func ServerHeartbeat(c *gin.Context) {
	var req struct {
		Id     server.IdKey  `json:"id"`
		Status com.RunStatus `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	servers := meta.GetTable().Servers()
	if err := servers.UpdateStatus(req.Id, req.Status); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c)
}

func ServerFetch(c *gin.Context) {
	servers := meta.GetTable().Servers()
	res, err := servers.FetchServerList()
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Data(c, map[string]interface{}{"list": res})
}
