package router

import (
	"example.com/micro-auth-demo/gateway/internal/handler"
	"example.com/micro-auth-demo/gateway/internal/middleware"
	"github.com/gin-gonic/gin"
)

func New() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	engine := gin.New()
	engine.Use(middleware.Recovery())

	engine.GET("/healthz", func(ctx *gin.Context) {
		ctx.String(200, "ok")
	})

	authGroup := engine.Group("/api/v1/auth")
	authGroup.POST("/login", handler.Login)
	authGroup.POST("/refresh", handler.Refresh)
	authGroup.Use(middleware.Auth())
	authGroup.POST("/logout", handler.Logout)
	authGroup.POST("/logout-all", handler.LogoutAll)
	authGroup.GET("/sessions", handler.ListSessions)
	authGroup.POST("/sessions/revoke", handler.RevokeSession)

	userGroup := engine.Group("/api/v1/users")
	userGroup.Use(middleware.Auth())
	userGroup.GET("/me", handler.Me)

	return engine
}
