package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/hackbeex/configcenter/server/model"
	"github.com/hackbeex/configcenter/util/response"
)

func CreateCluster(c *gin.Context) {
	var req model.CreateClusterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	cluster := model.ClusterModel{}
	res, err := cluster.Create(&req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Data(c, res)
}
