package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/hackbeex/configcenter/discover"
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
		response.JSON(c, nil, err)
		return
	}

	svr := server.Server{
		Id:   req.Id,
		Host: req.Host,
		Port: req.Port,
		Env:  req.Env,
	}
	if err := svr.Register(discover.GetStore()); err != nil {
		response.JSON(c, nil, err)
		return
	}
	response.JSON(c, nil, nil)
}

func ServerHeartbeat(c *gin.Context) {

}

func ServerFetch(c *gin.Context) {

}
