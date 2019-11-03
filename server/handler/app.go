package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/hackbeex/configcenter/server/model"
	"github.com/hackbeex/configcenter/util/response"
)

func GetAppList(c *gin.Context) {

}

func GetAppDetail(c *gin.Context) {

}

func CreateApp(c *gin.Context) {
	var req model.CreateAppReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	app := model.AppModel{}
	id, err := app.Create(&req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Data(c, map[string]string{
		"id": id,
	})
}
