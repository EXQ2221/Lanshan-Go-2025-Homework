package handler

import (
	"lesson10/internal/dto"
	"lesson10/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ToggleFavoriteHandler(favoriteSvc *service.FavoriteService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.FavorRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "format error",
			})
			return
		}

		uid := c.GetUint("user_id")
		if uid == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "please login",
			})
			return
		}

		isFavorited, err := favoriteSvc.ToggleFavoriteService(c.Request.Context(), uid, req.TargetType, req.TargetID)
		if err != nil {
			writeErr(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "success",
			"data":    gin.H{"is_favorited": *isFavorited}, // 操作后是否已收藏

		})
	}
}
