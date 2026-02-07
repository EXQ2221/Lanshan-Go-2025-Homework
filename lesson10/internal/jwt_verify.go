package internal

import (
	"lesson10/core"
	"lesson10/dao"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "authHeader is empty",
			})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "authHeader format error",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]
		token, err := core.ValidateToken(tokenString)
		if err != nil || !token.Valid {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "validate error",
			})
			c.Abort()
			return
		}

		claims := token.Claims.(jwt.MapClaims)

		userID := uint(claims["user_id"].(float64))
		username := claims["username"].(string)
		tokenVersion := int(claims["token_version"].(float64))
		roleF := claims["role"].(float64)
		role := core.Role(roleF)

		var user core.User
		if err := dao.DB.Select("id", "token_version").First(&user, userID).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			c.Abort()
			return
		}

		if user.TokenVersion != tokenVersion {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token expired"})
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Set("username", username)
		c.Set("token_version", tokenVersion)
		c.Set("role", role)
		c.Next()

	}
}
