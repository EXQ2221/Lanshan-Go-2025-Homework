package handler

import (
	"example.com/micro-auth-demo/gateway/internal/model"
	"github.com/gin-gonic/gin"
)

func writeJSON(ctx *gin.Context, status int, data any) {
	ctx.JSON(status, data)
}

func writeError(ctx *gin.Context, status int, message string) {
	writeJSON(ctx, status, model.APIResponse{
		Code:    status,
		Message: message,
	})
}
