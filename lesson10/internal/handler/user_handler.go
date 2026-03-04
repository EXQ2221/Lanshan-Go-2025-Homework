package handler

import (
	"fmt"
	"lesson10/internal/dto"
	"lesson10/internal/pkg/token"
	"lesson10/internal/service"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func RegisterHandler(userSvc *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "request format error",
			})
			return
		}

		user, err := userSvc.RegisterService(c.Request.Context(), req)

		if err != nil {
			writeErr(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":  "register success",
			"user_id":  user.ID,
			"username": user.Username,
		})
	}
}

func LoginHandler(userSvc *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "req format error",
			})
			return
		}

		tokenRes, user, err := userSvc.LoginService(c.Request.Context(), req)
		if err != nil {

			errMsg := err.Error()

			switch {
			case strings.Contains(errMsg, "username incorrect"):
				c.JSON(http.StatusUnauthorized, gin.H{
					"message": "username incorrect",
				})
				return

			case strings.Contains(errMsg, "password incorrect"):
				c.JSON(http.StatusUnauthorized, gin.H{
					"message": "password incorrect",
				})
				return

			default:
				writeErr(c, err)
				return
			}
		}

		refreshToken, err := token.GenerateRefreshToken(user.ID, user.TokenVersion)
		if err != nil {
			writeErr(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message":       "login success",
			"user_id":       user.ID,
			"username":      user.Username,
			"token":         tokenRes,
			"refresh_token": refreshToken,
		})

	}
}
func ChangePassHandler(userSvc *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.ChangePassRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "req format error",
			})
			return
		}

		id := c.GetUint("user_id")

		err := userSvc.ChangePassService(c.Request.Context(), req, id)

		if err != nil {
			writeErr(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"ok":            true,
			"need relog in": true,
		})
	}
}

func UpdateProfileHandler(userSvc *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.UpdateProfileRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "req format error",
			})
			return
		}

		id := c.GetUint("user_id")

		err := userSvc.UpdateProfileService(c.Request.Context(), req, id)
		if err != nil {
			writeErr(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"ok": true,
		})
	}
}

func UploadAvatarHandler(userSvc *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("user_id")

		file, err := c.FormFile("avatar")
		if err != nil {
			c.JSON(400, gin.H{"error": "missing avatar file"})
			return
		}

		// 1) 大小限制
		const maxSize = 5 * 1024 * 1024
		if file.Size > maxSize {
			c.JSON(400, gin.H{"error": "file too large (max 5MB)"})
			return
		}

		// 2) 打开读头部，检查 mime（防止随便传 .exe）
		f, err := file.Open()
		if err != nil {
			c.JSON(500, gin.H{"error": "open file failed"})
			return
		}
		defer f.Close()

		buf := make([]byte, 512)
		n, _ := f.Read(buf)
		contentType := http.DetectContentType(buf[:n])
		if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/webp" {
			c.JSON(400, gin.H{"error": "only jpg/png/webp allowed"})
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
			c.JSON(500, gin.H{"error": "mkdir failed"})
			return
		}

		savePath := filepath.Join(saveDir, filename)

		// 注意：刚刚读了512字节，不影响 SaveUploadedFile（它会重新打开文件）
		if err := c.SaveUploadedFile(file, savePath); err != nil {
			c.JSON(500, gin.H{"error": "save file failed"})
			return
		}

		// 5) 写库：avatar_url 存一个可访问的 url
		avatarURL := "/static/uploads/avatars/" + filename
		if err := userSvc.UpdateAvatarService(c.Request.Context(), userID, avatarURL); err != nil {
			c.JSON(500, gin.H{"error": "db update failed"})
			return
		}

		c.JSON(200, gin.H{"avatar_url": avatarURL})
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
			errMsg := err.Error()
			switch {
			case strings.Contains(errMsg, "fail to find user"):
				c.JSON(http.StatusNotFound, gin.H{
					"message": "fail to find user",
				})
				return

			default:
				log.Printf("error: %v", err)
				writeErr(c, err)
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "success",
			"data":    userPublicInfo,
		})
	}
}

func RefreshHandler(userSvc *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.RefreshRequest
		if err := c.ShouldBindJSON(&req); err != nil || req.RefreshToken == "" {
			c.JSON(400, gin.H{
				"error": "invalid refresh token",
			})
			return
		}

		tokenRes, err := token.ValidateToken(req.RefreshToken)

		if err != nil || !tokenRes.Valid {
			c.JSON(401, gin.H{"error": "invalid refresh token"})
			return
		}
		claims := tokenRes.Claims.(jwt.MapClaims)

		if claims["type"] != "refresh" {
			c.JSON(401, gin.H{"error": "invalid token type"})
			return
		}
		userID := uint(claims["user_id"].(float64))
		tokenVersion := int(claims["token_version"].(float64))

		newAccessToken, newRefreshToken, err := userSvc.Refresh(c.Request.Context(), userID, tokenVersion)
		if err != nil {
			writeErr(c, err)
			return
		}

		c.JSON(200, gin.H{
			"access_token":  newAccessToken,
			"refresh_token": newRefreshToken,
		})

	}
}
