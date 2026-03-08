package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Resp 统一响应结构体
// message: 提示信息
// code: HTTP 状态码
type Resp struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// JSON 统一响应输出
func JSON(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(code, Resp{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

// OK 快捷返回 200
func OK(c *gin.Context, data interface{}) {
	JSON(c, http.StatusOK, "success", data)
}

// Error 快捷返回错误
func Error(c *gin.Context, code int, message string) {
	JSON(c, code, message, nil)
}
