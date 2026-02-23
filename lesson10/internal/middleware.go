package internal

import (
	"lesson10/core"
	"lesson10/dao"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/time/rate"
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
			c.JSON(401, gin.H{
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

func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// 没有 token，直接放行，user_id = 0
			c.Set("user_id", uint(0))
			c.Next()
			return
		}

		// 有 token，正常解析
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Set("user_id", uint(0)) // 格式错误也当作未登录
			c.Next()
			return
		}

		tokenString := parts[1]
		token, err := core.ValidateToken(tokenString)
		if err != nil || !token.Valid {
			c.Set("user_id", uint(0)) // token 无效也放行
			c.Next()
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		userID := uint(claims["user_id"].(float64))

		// token_version 校验可以保留，但如果失败，仍然放行（只是 user_id=0）
		var user core.User
		if err := dao.DB.Select("token_version").First(&user, userID).Error; err != nil {
			c.Set("user_id", uint(0))
			c.Next()
			return
		}

		tokenVersion := int(claims["token_version"].(float64))
		if user.TokenVersion != tokenVersion {
			c.Set("user_id", uint(0)) // token 过期，当作未登录
			c.Next()
			return
		}

		c.Set("user_id", userID)
		c.Set("username", claims["username"].(string))
		c.Set("role", core.Role(claims["role"].(float64)))
		c.Next()
	}
}

// 每个 IP 的限流器
var visitors = make(map[string]*rate.Limiter)
var mu sync.Mutex

// 获取限流器
func getLimiter(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	limiter, exists := visitors[ip]
	if !exists {
		// 每秒 2 个请求，最多攒 5 个
		limiter = rate.NewLimiter(5, 10)
		visitors[ip] = limiter

		// 一个小时没访问就清理
		time.AfterFunc(60*time.Minute, func() {
			mu.Lock()
			delete(visitors, ip)
			mu.Unlock()
		})
	}
	return limiter
}

// 限流中间件
func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := getLimiter(ip)

		if !limiter.Allow() {
			c.JSON(429, gin.H{
				"error":       "too many requests",
				"retry_after": "1s",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
