package internal

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func InitRouter() {
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"}, // 前端端口
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour, // 预检缓存时间
	}))
	r.Static("/static", "./static")
	public := r.Group("/")
	{
		public.POST("/register", RegisterHandler)
		public.POST("/login", LoginHandler)

		public.GET("posts", ListPostsHandler)
		public.GET("/posts/comments", GetCommentsHandler)
		public.GET("/comments/:parent_id/replies", GetRepliesHandler)

		public.GET("/user/:id", GetUserInfoHandler)
		public.GET("/users/followers/:id", GetFollowersHandler) // 某用户的粉丝列表
		public.GET("/users/following/:id", GetFollowingHandler) // 某用户关注的人列表

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

		private.GET("/favorites", GetFavoritesHandler)
		private.GET("/draft", GetDraftHandler)
	}

	option := r.Group("/")
	option.Use(OptionalAuthMiddleware())
	{
		option.GET("/posts/:id", GetPostHandler)
	}
	r.Run(":8080")
}
