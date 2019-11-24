package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/hackbeex/configcenter/server/model"
	"github.com/hackbeex/configcenter/util/response"
)

func WatchConfig(c *gin.Context) {
	var req model.WatchConfigReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	config := model.ConfigModel{}
	res, err := config.Watch(&req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Data(c, res)
}

func GetClientConfigList(c *gin.Context) {
	var req model.ConfigListByAppReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	config := model.ConfigModel{}
	res, err := config.ListByApp(&req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Data(c, res)
}

func ExitClient(c *gin.Context) {
	var req model.ExitInstanceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	instance := model.InstanceModel{}
	err := instance.ExitInstance(&req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c)
}
