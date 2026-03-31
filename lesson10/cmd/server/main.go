package main

import (
	"fmt"
	"lesson10/internal/config"
	"lesson10/internal/model"
	"lesson10/internal/repository"
	"lesson10/internal/router"
	"lesson10/internal/service"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func loadEnv() {

	_ = godotenv.Overload(".env.local")

	_ = godotenv.Load(".env")

	if os.Getenv("DB_HOST") == "" {
		log.Fatal("DB_HOST is empty (check .env.local/.env)")
	}
}

func main() {

	loadEnv()

	_ = godotenv.Load(".env.local")
	config.InitDB()

	db := config.DB
	fmt.Println("开始迁移数据库...")
	err := db.AutoMigrate(
		&model.User{},
		&model.Post{},
		&model.Comment{},
		&model.PostImage{},
		&model.UserFollow{},
		&model.QuestionFollow{},
		&model.Reaction{},
		&model.Favorite{},
		&model.Activity{},
		&model.Notification{},
		&model.Conversation{},
		&model.ConversationMember{},
		&model.Message{},
		&model.Session{},
		&model.RefreshToken{},
		&model.SecurityEvent{},
	)

	if err != nil {
		log.Fatal("迁移失败: ", err)
	}
	fmt.Println("迁移完成")

	if err := config.CleanupPolymorphicTargetConstraints(); err != nil {
		log.Fatal("cleanup invalid polymorphic constraints failed: ", err)
	}

	userRepo := repository.NewUserRepo(db)
	postRepo := repository.NewPostRepo(db)
	commentRepo := repository.NewCommentRepo(db)
	notificationRepo := repository.NewNotificationRepo(db)
	reactionRepo := repository.NewReactionRepo(db)
	followRepo := repository.NewFollowRepo(db)
	favoriteRepo := repository.NewFavoriteRepo(db)
	sessionRepo := repository.NewSessionRepo(db)
	refreshTokenRepo := repository.NewRefreshTokenRepo(db)
	securityEventRepo := repository.NewSecurityEventRepo(db)

	userService := service.NewUserService(userRepo, followRepo, postRepo, db)
	authService := service.NewAuthService(userRepo, sessionRepo, refreshTokenRepo, securityEventRepo, db)
	userService.SetAuthService(authService)
	postService := service.NewPostService(userRepo, postRepo, favoriteRepo)
	commentService := service.NewCommentService(userRepo, postRepo, commentRepo, notificationRepo, reactionRepo)
	reactionService := service.NewReactionService(reactionRepo, postRepo, commentRepo, notificationRepo, db)
	followService := service.NewFollowService(followRepo, userRepo)
	favoriteService := service.NewFavoriteService(favoriteRepo, postRepo)
	notificationService := service.NewNotificationService(notificationRepo, userRepo)

	router.InitRouter(authService, userService, postService, commentService, reactionService, followService, favoriteService, notificationService)

}
