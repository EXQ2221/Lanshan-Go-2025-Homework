package handler

import (
	"lesson10/internal/dto"
	"lesson10/internal/pkg/response"
	"lesson10/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ToggleFavoriteHandler(favoriteSvc *service.FavoriteService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.FavorRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, "format error")
			return
		}

		uid := c.GetUint("user_id")
		if uid == 0 {
			response.Error(c, http.StatusBadRequest, "please login")
			return
		}

		isFavorited, err := favoriteSvc.ToggleFavoriteService(c.Request.Context(), uid, req.TargetType, req.TargetID)
		if err != nil {
			writeErr(c, err)
			return
		}

		response.OK(c, gin.H{
			"is_favorited": *isFavorited,
		})
	}
}
