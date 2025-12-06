package api

import (
	"lesson6/middleware"

	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	r := gin.Default()
	public := r.Group("/")
	{
		public.POST("/login", Login)
		public.POST("/register", Register)
		public.POST("/refresh", Refresh)
	}

	private := r.Group("/")
	private.Use(middleware.AuthMiddleware())
	{
		private.POST("/changepassword", ChangePassword)
	}
	return r
}
