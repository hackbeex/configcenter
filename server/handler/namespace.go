package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/hackbeex/configcenter/server/model"
	"github.com/hackbeex/configcenter/util/response"
)

func CreateNamespace(c *gin.Context) {
	var req model.CreateNamespaceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	namespace := model.NamespaceModel{}
	res, err := namespace.Create(&req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Data(c, res)
}
