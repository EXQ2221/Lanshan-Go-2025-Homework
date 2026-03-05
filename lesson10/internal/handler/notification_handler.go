package handler

import (
	"errors"
	"lesson10/internal/pkg/errcode"
	"lesson10/internal/service"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func writeErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, errcode.ErrBadRequest):
		c.JSON(400, gin.H{"error": "bad_request"})
	case errors.Is(err, errcode.ErrUnauthorized):
		c.JSON(401, gin.H{"error": "unauthorized"})
	case errors.Is(err, errcode.ErrForbidden):
		c.JSON(403, gin.H{"error": "forbidden"})
	case errors.Is(err, errcode.ErrConflict):
		c.JSON(409, gin.H{"error": "conflict"})
	default:
		log.Println("internal error:", err)
		c.JSON(500, gin.H{"error": "server_error"})
	}
}

func GetNotificationsHandler(notificationSvc *service.NotificationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetUint("user_id")
		if uid == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "please login first",
			})
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

		unreadOnly := c.DefaultQuery("unread_only", "0") == "1"

		notifications, total, err := notificationSvc.GetNotifications(c.Request.Context(), uid, page, size, unreadOnly)
		if err != nil {
			log.Printf("get notifications failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "success",
			"data": gin.H{
				"notifications": notifications,
				"total":         total,
				"page":          page,
				"size":          size,
			},
		})
	}
}

func GetUnreadCountHandler(notificationSvc *service.NotificationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetUint("user_id")

		count, err := notificationSvc.GetUnreadCountService(c.Request.Context(), uid)
		if err != nil {
			writeErr(c, err)
			return
		}

		c.JSON(200, gin.H{"count": count})
	}
}

func MarkAllNotificationsReadHandler(notificationSvc *service.NotificationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetUint("user_id")
		if uid == 0 {
			c.JSON(401, gin.H{"message": "please login first"})
			return
		}

		if err := notificationSvc.MarkAllNotificationsRead(c.Request.Context(), uid); err != nil {
			writeErr(c, err)
			return
		}

		c.JSON(200, gin.H{"message": "success"})
	}
}
