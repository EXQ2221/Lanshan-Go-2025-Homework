package handler

import (
	"fmt"
	"lesson10/internal/dto"
	"lesson10/internal/pkg/response"
	"lesson10/internal/pkg/token"
	"lesson10/internal/service"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
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

func LoginHandler(userSvc *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, "req format error")
			return
		}

		accessToken, refreshToken, user, err := userSvc.LoginService(
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
			"user_id":       user.ID,
			"username":      user.Username,
			"token":         accessToken,
			"refresh_token": refreshToken,
		})
	}
}

func ChangePassHandler(userSvc *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.ChangePassRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, "req format error")
			return
		}

		id := c.GetUint("user_id")

		err := userSvc.ChangePassService(c.Request.Context(), req, id)

		if err != nil {
			writeErr(c, err)
			return
		}

		response.OK(c, gin.H{
			"ok":            true,
			"need relog in": true,
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

		id := c.GetUint("user_id")

		err := userSvc.UpdateProfileService(c.Request.Context(), req, id)
		if err != nil {
			writeErr(c, err)
			return
		}

		response.OK(c, gin.H{
			"ok": true,
		})
	}
}

func UploadAvatarHandler(userSvc *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("user_id")

		file, err := c.FormFile("avatar")
		if err != nil {
			response.Error(c, 400, "missing avatar file")
			return
		}

		// 1) 大小限制
		const maxSize = 5 * 1024 * 1024
		if file.Size > maxSize {
			response.Error(c, 400, "file too large (max 5MB)")
			return
		}

		// 2) 打开读头部，检查 mime（防止随便传 .exe）
		f, err := file.Open()
		if err != nil {
			response.Error(c, 500, "open file failed")
			return
		}
		defer f.Close()

		buf := make([]byte, 512)
		n, _ := f.Read(buf)
		contentType := http.DetectContentType(buf[:n])
		if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/webp" {
			response.Error(c, 400, "only jpg/png/webp allowed")
			return
		}

		// 3) 生成文件名（避免重名/路径穿越）
		ext := ".jpg"
		switch contentType {
		case "image/png":
			ext = ".png"
		case "image/webp":
			ext = ".webp"
		}

		// 你也可以用 uuid，这里用 时间戳+userID
		filename := fmt.Sprintf("u%d_%d%s", userID, time.Now().UnixNano(), ext)

		// 4) 确保目录存在
		saveDir := "static/uploads/avatars"
		if err := os.MkdirAll(saveDir, 0755); err != nil {
			response.Error(c, 500, "mkdir failed")
			return
		}

		savePath := filepath.Join(saveDir, filename)

		// 注意：刚刚读了512字节，不影响 SaveUploadedFile（它会重新打开文件）
		if err := c.SaveUploadedFile(file, savePath); err != nil {
			response.Error(c, 500, "save file failed")
			return
		}

		// 5) 写库：avatar_url 存一个可访问的 url
		avatarURL := "/static/uploads/avatars/" + filename
		if err := userSvc.UpdateAvatarService(c.Request.Context(), userID, avatarURL); err != nil {
			response.Error(c, 500, "db update failed")
			return
		}

		response.OK(c, gin.H{"avatar_url": avatarURL})
	}
}

func GetUserInfoHandler(userSvc *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.Param("id")
		userIDUint64, err := strconv.ParseUint(userIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "id format error",
			})
			return
		}

		page := 1
		pageStr := c.DefaultQuery("page", "1")
		if p, err := strconv.Atoi(pageStr); err == nil && p >= 1 {
			page = p
		}

		currentID := c.GetUint("user_id")
		userPublicInfo, err := userSvc.GetUserInfoService(c.Request.Context(), currentID, uint(userIDUint64), page)
		if err != nil {
			writeErr(c, err)
			return
		}
		response.OK(c, userPublicInfo)
	}
}

func RefreshHandler(userSvc *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.RefreshRequest
		if err := c.ShouldBindJSON(&req); err != nil || req.RefreshToken == "" {
			response.Error(c, 400, "invalid refresh token")
			return
		}

		tk, err := token.ValidateToken(req.RefreshToken)
		if err != nil || !tk.Valid {
			response.Error(c, 401, "invalid refresh token")
			return
		}

		claims, ok := tk.Claims.(jwt.MapClaims)
		if !ok {
			response.Error(c, 401, "invalid claims")
			return
		}

		if claims["type"] != "refresh" {
			response.Error(c, 401, "invalid token type")
			return
		}

		userIDf, ok1 := claims["user_id"].(float64)
		tokenVerf, ok2 := claims["token_version"].(float64)
		sid, ok3 := claims["sid"].(string)

		if !ok1 || !ok2 || !ok3 || sid == "" {
			response.Error(c, 401, "invalid refresh claims")
			return
		}

		ip := c.ClientIP()
		ua := c.GetHeader("User-Agent")

		access, refresh, needRelogin, err := userSvc.RefreshWithWhitelist(
			c.Request.Context(),
			req.RefreshToken,
			uint(userIDf),
			int(tokenVerf),
			sid,
			ip,
			ua,
		)

		if err != nil {
			writeErr(c, err)
			return
		}
		resp := gin.H{
			"access_token":  access,
			"refresh_token": refresh,
			"need_relogin":  needRelogin,
		}
		msg := "success"
		if needRelogin {
			msg = "检测到异地登录，请重新登录"
		}
		response.JSON(c, 200, msg, resp)
	}
}
