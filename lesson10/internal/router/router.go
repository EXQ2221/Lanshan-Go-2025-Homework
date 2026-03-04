package router

import (
	"lesson10/internal/handler"
	"lesson10/internal/middleware"
	"lesson10/internal/service"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func InitRouter(userService *service.UserService,
	postService *service.PostService,
	commentService *service.CommentService,
	reactionService *service.ReactionService,
	followService *service.FollowService,
	favoriteService *service.FavoriteService,
	notification *service.NotificationService) {
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
	public.Use(middleware.RateLimit())
	{
		public.POST("/register", handler.RegisterHandler(userService))
		public.POST("/login", handler.LoginHandler(userService))

		public.GET("posts", handler.ListPostsHandler)
		public.GET("/posts/comments", handler.GetCommentsHandler(commentService))
		public.GET("/comments/:parent_id/replies", handler.GetRepliesHandler(commentService))

		public.GET("/users/followers/:id", handler.GetFollowersHandler(followService)) // 某用户的粉丝列表
		public.GET("/users/following/:id", handler.GetFollowingHandler(followService)) // 某用户关注的人列表

	}

	private := r.Group("/")
	private.Use(middleware.AuthMiddleware())
	private.Use(middleware.RateLimit())
	{
		private.PUT("/change_pass", handler.ChangePassHandler(userService))
		private.PUT("/profile", handler.UpdateProfileHandler(userService))
		private.POST("/avatar", handler.UploadAvatarHandler(userService))

		private.POST("/posts", handler.CreatePostHandler(postService))
		private.PUT("/posts/:id", handler.UpdatePostHandler(postService))
		private.DELETE("posts/:id", handler.DeletePostHandler(postService))

		private.POST("/comments", handler.PostCommentHandler(commentService))
		private.DELETE("/comments/:id", handler.DeleteCommentHandler(commentService))

		private.POST("follow/:id", handler.FollowUserHandler(followService))
		private.DELETE("/follow/:id", handler.UnfollowUserHandler(followService))

		private.POST("/upload/article-image", handler.UploadArticleImageHandler)

		private.POST("/reactions", handler.ToggleReactionHandler(reactionService)) //点赞
		private.POST("/favorites", handler.ToggleFavoriteHandler(favoriteService)) //收藏

		private.GET("/notifications", handler.GetNotificationsHandler(notification))

		private.GET("/favorites", handler.GetFavoritesHandler(postService))
		private.GET("/draft", handler.GetDraftHandler(postService))
		private.GET("/notifications/count", handler.GetUnreadCountHandler(notification))

		private.POST("/notifications/read-all", handler.MarkAllNotificationsReadHandler(notification))
	}

	option := r.Group("/")
	option.Use(middleware.OptionalAuthMiddleware())
	option.Use(middleware.RateLimit())
	{
		option.GET("/posts/:id", handler.GetPostHandler(postService))
		option.POST("/refresh", handler.RefreshHandler(userService))
		option.GET("/user/:id", handler.GetUserInfoHandler(userService))
	}
	r.Run(":8080")
}
