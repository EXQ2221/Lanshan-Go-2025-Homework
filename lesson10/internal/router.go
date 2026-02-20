package internal

import "github.com/gin-gonic/gin"

func InitRouter() {
	r := gin.Default()
	r.Static("/static", "./static")
	public := r.Group("/")
	{
		public.POST("/register", RegisterHandler)
		public.POST("/login", LoginHandler)

		public.GET("posts", ListPostsHandler)
		public.GET("/posts/:id", GetPostHandler)
		public.GET("/posts/comments", GetCommentsHandler)
		public.GET("/comments/:comment_id/replies", GetRepliesHandler)

		public.GET("/user/:id", GetUserInfoHandler)
		public.GET("/users/:id/followers", GetFollowersHandler) // 某用户的粉丝列表
		public.GET("/users/:id/following", GetFollowingHandler) // 某用户关注的人列表

	}

	private := r.Group("/")
	private.Use(AuthMiddleware())
	{
		private.PUT("/change_pass", ChangePassHandler)
		private.PUT("/profile", UpdateProfileHandler)
		private.POST("/avatar", UploadAvatarHandler)

		private.POST("/posts", CreatePostHandler)
		private.PUT("/posts/:id", UpdatePostHandler)
		private.DELETE("posts/:id", DeletePostHandler)

		private.POST("/comments", PostCommentHandler)
		private.DELETE("/comments/:id", DeleteCommentHandler)

		private.POST("follow/:id", FollowUserHandler)
		private.DELETE("/follow/:id", UnfollowUserHandler)

		private.POST("/upload/article-image", UploadArticleImageHandler)

		private.POST("/reactions", ToggleReactionHandler) //点赞
		private.POST("/favorites", ToggleFavoriteHandler) //收藏

		private.GET("/notifications", GetNotificationsHandler)
	}
	r.Run(":8080")
}
