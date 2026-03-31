package handler

import (
	"errors"
	"lesson10/internal/pkg/errcode"
	"lesson10/internal/pkg/response"
	"lesson10/internal/service"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func writeErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, errcode.ErrUsernameIncorrect):
		response.Error(c, 401, errcode.ErrUsernameIncorrect.Error())
	case errors.Is(err, errcode.ErrPasswordIncorrect):
		response.Error(c, 401, errcode.ErrPasswordIncorrect.Error())
	case errors.Is(err, errcode.ErrSessionRevoked):
		response.Error(c, 401, errcode.ErrSessionRevoked.Error())
	case errors.Is(err, errcode.ErrRefreshReuse):
		response.Error(c, 403, errcode.ErrRefreshReuse.Error())
	case errors.Is(err, errcode.ErrDeviceMismatch):
		response.Error(c, 401, errcode.ErrDeviceMismatch.Error())
	case errors.Is(err, errcode.ErrHasFollowed):
		response.Error(c, 400, errcode.ErrHasFollowed.Error())
	case errors.Is(err, errcode.ErrHasNotFollowed):
		response.Error(c, 400, errcode.ErrHasNotFollowed.Error())
	case errors.Is(err, errcode.ErrInvalidListType):
		response.Error(c, 400, errcode.ErrInvalidListType.Error())
	case errors.Is(err, errcode.ErrBadRequest):
		response.Error(c, 400, errcode.ErrBadRequest.Error())
	case errors.Is(err, errcode.ErrUnauthorized):
		response.Error(c, 401, errcode.ErrUnauthorized.Error())
	case errors.Is(err, errcode.ErrForbidden):
		response.Error(c, 403, errcode.ErrForbidden.Error())
	case errors.Is(err, errcode.ErrConflict):
		response.Error(c, 409, errcode.ErrConflict.Error())
	case errors.Is(err, errcode.ErrNotFound):
		response.Error(c, 404, errcode.ErrNotFound.Error())
	case errors.Is(err, errcode.ErrInternal):
		response.Error(c, 500, errcode.ErrInternal.Error())
	default:
		response.Error(c, 500, err.Error())
	}
}

func GetNotificationsHandler(notificationSvc *service.NotificationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetUint("user_id")
		if uid == 0 {
			response.Error(c, http.StatusUnauthorized, "please login first")
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
			response.Error(c, http.StatusInternalServerError, "internal server error")
			return
		}

		response.OK(c, gin.H{
			"notifications": notifications,
			"total":         total,
			"page":          page,
			"size":          size,
		})
	}
}

func GetUnreadCountHandler(notificationSvc *service.NotificationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetUint("user_id")
		if uid == 0 {
			response.Error(c, http.StatusUnauthorized, "please login first")
			return
		}

		count, err := notificationSvc.GetUnreadCountService(c.Request.Context(), uid)
		if err != nil {
			writeErr(c, err)
			return
		}

		response.OK(c, gin.H{"count": count})
	}
}

func MarkAllNotificationsReadHandler(notificationSvc *service.NotificationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetUint("user_id")
		if uid == 0 {
			response.Error(c, http.StatusUnauthorized, "please login first")
			return
		}

		if err := notificationSvc.MarkAllNotificationsRead(c.Request.Context(), uid); err != nil {
			writeErr(c, err)
			return
		}

		response.JSON(c, http.StatusOK, "success", nil)
	}
}
