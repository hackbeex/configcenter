package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/hackbeex/configcenter/discover"
	"github.com/hackbeex/configcenter/discover/client"
	"github.com/hackbeex/configcenter/util/response"
)

func ClientRegister(c *gin.Context) {
	var req struct {
		AppId   string
		Host    string
		Port    int
		Cluster string
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.JSON(c, nil, err)
		return
	}

	clt := client.Client{
		AppId:   req.AppId,
		Host:    req.Host,
		Port:    req.Port,
		Cluster: req.Cluster,
	}
	if err := clt.Register(discover.GetStore()); err != nil {
		response.JSON(c, nil, err)
		return
	}
	response.JSON(c, nil, nil)
}

func ClientHeartbeat(c *gin.Context) {

}

func ClientFetch(c *gin.Context) {

}
