package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/hackbeex/configcenter/discover"
	"github.com/hackbeex/configcenter/discover/client"
	"github.com/hackbeex/configcenter/discover/com"
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
	var req struct {
		AppId client.AppIdKey
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	clients := discover.GetTable().Clients()
	if err := clients.UpdateStatus(req.AppId, com.OnlineStatus); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c)
}

func ClientFetch(c *gin.Context) {
	clients := discover.GetTable().Clients()
	res, err := clients.FetchClientList()
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Data(c, map[string]interface{}{"list": res})
}
