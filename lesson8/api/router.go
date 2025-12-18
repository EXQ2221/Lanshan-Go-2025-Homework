package api

import (
	"lesson8/middleware"

	"github.com/gin-gonic/gin"
)

func InitRouter() {
	r := gin.Default()

	public := r.Group("/")
	{
		public.POST("/register", RegisterHandler)
		public.POST("/login", LoginHandler)
	}

	private := r.Group("/")
	private.Use(middleware.AuthMiddleware())
	{
		private.GET("/todo", GetTodosHandler)
		private.POST("/todo", CreateTodoHandler)
		private.PUT("/todo/:id", UpdateTodoHandler)
		private.DELETE("todo/:id", DeleteTodoHandler)
	}

	r.Run(":8080")
}
