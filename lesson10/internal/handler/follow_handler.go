package handler

import (
	"lesson10/internal/service"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func FollowUserHandler(followSvc *service.FollowService) gin.HandlerFunc {
	return func(c *gin.Context) {
		followeeIDStr := c.Param("id")
		followeeID, err := strconv.ParseUint(followeeIDStr, 10, 64)
		if err != nil || followeeID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "id format incorrect",
			})
			return
		}

		followerID := c.GetUint("user_id")
		if followerID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "please log in",
			})
			return
		}

		if followerID == uint(followeeID) {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "can not follow yourself",
			})
			return
		}

		err = followSvc.FollowUserService(c.Request.Context(), followerID, uint(followeeID))

		if err != nil {
			errMsg := err.Error()
			switch {
			case strings.Contains(errMsg, "has followed"):
				c.JSON(http.StatusBadRequest, gin.H{
					"message": "has followed",
				})
				return
			default:
				log.Printf("fail: %v", err)
				writeErr(c, err)
				return
			}

		}

		c.JSON(http.StatusOK, gin.H{
			"message": "success",
		})
	}
}

func UnfollowUserHandler(followSvc *service.FollowService) gin.HandlerFunc {
	return func(c *gin.Context) {
		followeeIDStr := c.Param("id")
		followeeID, err := strconv.ParseUint(followeeIDStr, 10, 64)

		if err != nil || followeeID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "id format incorrect",
			})
		}

		followerID := c.GetUint("user_id")
		if followerID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "please log in",
			})
		}

		err = followSvc.UnfollowUserService(c.Request.Context(), followerID, uint(followeeID))
		if err != nil {
			errMsg := err.Error()
			if strings.Contains(errMsg, "has not followed") {
				c.JSON(http.StatusBadRequest, gin.H{
					"message": "has not followed",
				})
				return

			} else {
				writeErr(c, err)
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "success",
		})
	}
}

func GetFollowersHandler(followSvc *service.FollowService) gin.HandlerFunc {
	return func(c *gin.Context) {
		getFollowListInternal(c, followSvc, "followers")
	}
}

func GetFollowingHandler(followSvc *service.FollowService) gin.HandlerFunc {
	return func(c *gin.Context) {
		getFollowListInternal(c, followSvc, "following")
	}
}

// 内部共用函数（不暴露给 Gin）
func getFollowListInternal(c *gin.Context, followSvc *service.FollowService, listType string) {
	targetUserIDStr := c.Param("id")
	targetUserID, err := strconv.ParseUint(targetUserIDStr, 10, 64)
	if err != nil || targetUserID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid user id"})
		return
	}

	pageStr := c.DefaultQuery("page", "1")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}

	sizeStr := c.DefaultQuery("size", "20")
	size, _ := strconv.Atoi(sizeStr)
	if size < 1 || size > 50 {
		size = 20
	}

	currentUserID := c.GetUint("user_id") // 当前登录用户（用于 is_followed，可选）

	users, total, err := followSvc.GetFollowListService(c.Request.Context(), uint(targetUserID), listType, currentUserID, page, size)
	if err != nil {
		log.Printf("get follow list failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"data": gin.H{
			"users": users,
			"total": total,
			"page":  page,
			"size":  size,
		},
	})
}
