package middleware

import (
	"lesson6/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

var secretKey = []byte("hfhiwii938rjnejnaoef")

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "can not find authheader",
			})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")

		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "class is not token",
			})
			c.Abort()
			return
		}
		tokenString := parts[1]
		token, err := utils.ValidateToken(tokenString)
		if err != nil || !token.Valid {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "token incorrect",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
