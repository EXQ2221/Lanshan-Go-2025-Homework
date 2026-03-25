package handler

import (
	"example.com/micro-auth-demo/gateway/internal/middleware"
	"example.com/micro-auth-demo/gateway/internal/model"
	"example.com/micro-auth-demo/gateway/internal/rpc"
	userpb "example.com/micro-auth-demo/gateway/kitex_gen/user"
	"github.com/gin-gonic/gin"
)

func Me(ctx *gin.Context) {
	authCtx, ok := middleware.GetAuthContext(ctx)
	if !ok {
		writeError(ctx, 401, "missing auth context")
		return
	}

	client, err := rpc.UserClient()
	if err != nil {
		writeError(ctx, 500, err.Error())
		return
	}

	userResp, err := client.GetUser(ctx.Request.Context(), &userpb.GetUserRequest{
		UserId: authCtx.UserID,
	})
	if err != nil {
		writeError(ctx, 500, err.Error())
		return
	}

	writeJSON(ctx, 200, model.APIResponse{
		Code:    0,
		Message: "success",
		Data: model.UserInfo{
			UserID:   userResp.User.UserId,
			Email:    userResp.User.Email,
			Nickname: userResp.User.Nickname,
		},
	})
}
