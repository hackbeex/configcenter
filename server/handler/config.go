package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/hackbeex/configcenter/server/model"
	"github.com/hackbeex/configcenter/util/response"
)

func GetConfigDetail(c *gin.Context) {
	var req model.ConfigDetailReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	config := model.ConfigModel{}
	res, err := config.Detail(&req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Data(c, res)
}

func CreateConfig(c *gin.Context) {
	var req model.CreateConfigReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	config := model.ConfigModel{}
	res, err := config.Create(&req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Data(c, res)
}

func UpdateConfig(c *gin.Context) {
	var req model.UpdateConfigReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	config := model.ConfigModel{}
	err := config.Update(&req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c)
}

func DeleteConfig(c *gin.Context) {
	var req model.DeleteConfigReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	config := model.ConfigModel{}
	err := config.Delete(&req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c)
}

func ReleaseConfig(c *gin.Context) {

}

func RollbackConfig(c *gin.Context) {

}

func SyncConfig(c *gin.Context) {

}

func GetConfigHistory(c *gin.Context) {

}

func WatchConfig(c *gin.Context) {

}
