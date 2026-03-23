package handler

import (
	"lesson10/internal/dto"
	"lesson10/internal/pkg/response"
	"lesson10/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ToggleReactionHandler(reactionSvc *service.ReactionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.LikeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, "format error")
			return
		}

		uid := c.GetUint("user_id")
		if uid == 0 {
			response.Error(c, http.StatusUnauthorized, "please login")
			return
		}

		isLiked, err := reactionSvc.ToggleReactionService(c.Request.Context(), uid, req.TargetType, req.TargetID)
		if err != nil {
			writeErr(c, err)
			return
		}

		response.OK(c, gin.H{"status": isLiked})
	}
}
