package internal

import "github.com/gin-gonic/gin"

func InitRouter() {
	r := gin.Default()

	public := r.Group("/")
	{
		public.POST("/register", RegisterHandler)
		public.POST("/login", LoginHandler)
		public.GET("posts", ListPostsHandler)
		public.GET("/posts/:id", GetPostHandler)
		public.GET("/posts/:id/comments", GetCommentsHandler)

	}

	private := r.Group("/")
	private.Use(AuthMiddleware())
	{
		private.PUT("/change_pass", ChangePassHandler)
		private.PUT("/profile", UpdateProfileHandler)
		private.POST("/avatar", UploadAvatarHandler)
		private.POST("/posts", CreatePostHandler)
		private.POST("/comments", PostCommentHandler)
		private.PUT("/posts/:id")
		private.DELETE("posts/:id")
	}
	r.Run(":8080")
}
