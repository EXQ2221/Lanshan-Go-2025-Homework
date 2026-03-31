package handler

import (
	"fmt"
	"lesson10/internal/dto"
	"lesson10/internal/pkg/response"
	"lesson10/internal/service"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func RegisterHandler(userSvc *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, "request format error")
			return
		}

		user, err := userSvc.RegisterService(c.Request.Context(), req)
		if err != nil {
			writeErr(c, err)
			return
		}

		response.OK(c, gin.H{
			"user_id":  user.ID,
			"username": user.Username,
		})
	}
}

func LoginHandler(authSvc *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, "req format error")
			return
		}

		pair, user, deviceID, err := authSvc.Login(
			c.Request.Context(),
			req,
			c.ClientIP(),
			c.GetHeader("User-Agent"),
		)
		if err != nil {
			writeErr(c, err)
			return
		}

		response.OK(c, gin.H{
			"user_id":            user.ID,
			"username":           user.Username,
			"token":              pair.AccessToken,
			"refresh_token":      pair.RefreshToken,
			"session_id":         pair.SessionId,
			"device_id":          deviceID,
			"access_expires_at":  pair.AccessExpiresAt,
			"refresh_expires_at": pair.RefreshExpiresAt,
		})
	}
}

func RefreshHandler(authSvc *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.RefreshRequest
		if err := c.ShouldBindJSON(&req); err != nil || req.RefreshToken == "" {
			response.Error(c, http.StatusBadRequest, "invalid refresh token")
			return
		}

		pair, err := authSvc.Refresh(c.Request.Context(), req, c.ClientIP(), c.GetHeader("User-Agent"))
		if err != nil {
			writeErr(c, err)
			return
		}

		response.OK(c, gin.H{
			"access_token":       pair.AccessToken,
			"refresh_token":      pair.RefreshToken,
			"session_id":         pair.SessionId,
			"access_expires_at":  pair.AccessExpiresAt,
			"refresh_expires_at": pair.RefreshExpiresAt,
		})
	}
}

func LogoutHandler(authSvc *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken := c.GetString("access_token")
		if accessToken == "" {
			response.Error(c, http.StatusUnauthorized, "missing access token")
			return
		}

		if err := authSvc.Logout(c.Request.Context(), accessToken); err != nil {
			writeErr(c, err)
			return
		}

		response.OK(c, gin.H{"ok": true})
	}
}

func LogoutAllHandler(authSvc *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.LogoutAllRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, "req format error")
			return
		}

		if err := authSvc.LogoutAll(c.Request.Context(), c.GetUint("user_id"), req.Password); err != nil {
			writeErr(c, err)
			return
		}

		response.OK(c, gin.H{"ok": true})
	}
}

func ListSessionsHandler(authSvc *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessions, err := authSvc.ListSessions(c.Request.Context(), c.GetUint("user_id"), c.GetString("session_id"))
		if err != nil {
			writeErr(c, err)
			return
		}

		response.OK(c, gin.H{"sessions": sessions})
	}
}

func RevokeSessionHandler(authSvc *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.RevokeSessionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, "req format error")
			return
		}

		if err := authSvc.RevokeSession(c.Request.Context(), c.GetUint("user_id"), req.SessionID, req.Password); err != nil {
			writeErr(c, err)
			return
		}

		response.OK(c, gin.H{"ok": true})
	}
}

func ChangePassHandler(userSvc *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.ChangePassRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, "req format error")
			return
		}

		if err := userSvc.ChangePassService(c.Request.Context(), req, c.GetUint("user_id")); err != nil {
			writeErr(c, err)
			return
		}

		response.OK(c, gin.H{
			"ok":             true,
			"need_relogin":   true,
			"sessions_reset": true,
		})
	}
}

func UpdateProfileHandler(userSvc *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.UpdateProfileRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, "req format error")
			return
		}

		if err := userSvc.UpdateProfileService(c.Request.Context(), req, c.GetUint("user_id")); err != nil {
			writeErr(c, err)
			return
		}

		response.OK(c, gin.H{"ok": true})
	}
}

func UploadAvatarHandler(userSvc *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("user_id")

		file, err := c.FormFile("avatar")
		if err != nil {
			response.Error(c, http.StatusBadRequest, "missing avatar file")
			return
		}

		const maxSize = 5 * 1024 * 1024
		if file.Size > maxSize {
			response.Error(c, http.StatusBadRequest, "file too large (max 5MB)")
			return
		}

		f, err := file.Open()
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "open file failed")
			return
		}
		defer f.Close()

		buf := make([]byte, 512)
		n, _ := f.Read(buf)
		contentType := http.DetectContentType(buf[:n])
		if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/webp" {
			response.Error(c, http.StatusBadRequest, "only jpg/png/webp allowed")
			return
		}

		ext := ".jpg"
		switch contentType {
		case "image/png":
			ext = ".png"
		case "image/webp":
			ext = ".webp"
		}

		filename := fmt.Sprintf("u%d_%d%s", userID, time.Now().UnixNano(), ext)
		saveDir := "static/uploads/avatars"
		if err := os.MkdirAll(saveDir, 0o755); err != nil {
			response.Error(c, http.StatusInternalServerError, "mkdir failed")
			return
		}

		savePath := filepath.Join(saveDir, filename)
		if err := c.SaveUploadedFile(file, savePath); err != nil {
			response.Error(c, http.StatusInternalServerError, "save file failed")
			return
		}

		avatarURL := "/static/uploads/avatars/" + filename
		if err := userSvc.UpdateAvatarService(c.Request.Context(), userID, avatarURL); err != nil {
			response.Error(c, http.StatusInternalServerError, "db update failed")
			return
		}

		response.OK(c, gin.H{"avatar_url": avatarURL})
	}
}

func GetUserInfoHandler(userSvc *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDUint64, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "id format error")
			return
		}

		page := 1
		if p, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil && p >= 1 {
			page = p
		}

		userPublicInfo, err := userSvc.GetUserInfoService(c.Request.Context(), c.GetUint("user_id"), uint(userIDUint64), page)
		if err != nil {
			writeErr(c, err)
			return
		}

		response.OK(c, userPublicInfo)
	}
}
