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

func GetConfigList(c *gin.Context) {
	var req model.ConfigListReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	config := model.ConfigModel{}
	res, err := config.List(&req)
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

func GetConfigHistory(c *gin.Context) {
	var req model.ConfigHistoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	config := model.ConfigModel{}
	res, err := config.GetHistory(&req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Data(c, res)
}

func ReleaseConfig(c *gin.Context) {
	var req model.ReleaseConfigReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	config := model.ConfigModel{}
	err := config.Release(&req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c)
}

func GetConfigReleaseHistory(c *gin.Context) {
	var req model.ConfigReleaseHistoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	config := model.ConfigModel{}
	res, err := config.GetReleaseHistory(&req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Data(c, res)
}

func RollbackConfig(c *gin.Context) {
	var req model.RollbackConfigReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	config := model.ConfigModel{}
	err := config.Rollback(&req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c)
}

func SyncConfig(c *gin.Context) {
	var req model.SyncConfigReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	config := model.ConfigModel{}
	err := config.Sync(&req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c)
}

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
