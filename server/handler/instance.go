package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/hackbeex/configcenter/server/model"
	"github.com/hackbeex/configcenter/util/response"
)

func GetInstanceList(c *gin.Context) {
	var req model.InstanceListReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	instance := model.InstanceModel{}
	res, err := instance.List(&req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Data(c, res)
}

func ExitInstance(c *gin.Context) {
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
