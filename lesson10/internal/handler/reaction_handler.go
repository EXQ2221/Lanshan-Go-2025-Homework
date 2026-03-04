package handler

import (
	"lesson10/internal/dto"
	"lesson10/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ToggleReactionHandler(reactionSvc *service.ReactionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.LikeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "format error",
			})
			return
		}

		uid := c.GetUint("user_id")
		if uid == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "please login",
			})
			return
		}

		isLiked, err := reactionSvc.ToggleReactionService(c.Request.Context(), uid, req.TargetType, req.TargetID)
		if err != nil {
			writeErr(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "success",
			"status":  isLiked,
		})
	}
}
