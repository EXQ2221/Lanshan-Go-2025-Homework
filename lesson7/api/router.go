package api

import (
	"lesson7/middleware"

	"github.com/gin-gonic/gin"
)

func InitRouter() {
	r := gin.Default()

	public := r.Group("/")
	{
		public.POST("/register", Register)
		public.POST("/login", Login)
	}

	private := r.Group("/")
	private.Use(middleware.AuthMiddleware())
	{
		private.GET("/todo", GetTodos)
		private.POST("/todo", CreateTodo)
		private.PUT("/todo/:id", UpdateTodo)
		private.DELETE("todo/:id", DeleteTodo)
	}

	r.Run(":8080")
}
