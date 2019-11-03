package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/hackbeex/configcenter/server/model"
	"github.com/hackbeex/configcenter/util/response"
)

func GetAppList(c *gin.Context) {
	var req model.AppListReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	app := model.AppModel{}
	res, err := app.List(&req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Data(c, res)
}

func GetAppDetail(c *gin.Context) {
	var req model.AppDetailReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	app := model.AppModel{}
	res, err := app.Detail(&req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Data(c, res)
}

func CreateApp(c *gin.Context) {
	var req model.CreateAppReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	app := model.AppModel{}
	res, err := app.Create(&req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Data(c, res)
}
