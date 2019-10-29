package response

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Result struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func JSON(c *gin.Context, data interface{}, err error) {
	code := http.StatusOK
	msg := ""
	if err != nil {
		code = http.StatusBadRequest
		msg = err.Error()
	}
	if data == nil {
		data = map[string]string{}
	}
	c.JSON(http.StatusOK, Result{
		Code:    code,
		Message: msg,
		Data:    data,
	})
}
