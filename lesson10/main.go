package main

import (
	"fmt"
	"lesson10/core"
	"lesson10/dao"
	"lesson10/internal"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func TestDeleteFavorite() {
	err := dao.DB.Where("user_id = 3 AND target_type = 1 AND target_id = 3").
		Delete(&core.Favorite{}).Error
	log.Printf("test delete err: %v", err)
	log.Printf("test rows affected: %d", dao.DB.RowsAffected)
}
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
	dao.InitDB()

	internal.InitRouter()

	db := dao.DB

	fmt.Println("开始迁移数据库...")
	err := db.AutoMigrate(&core.User{})

	if err != nil {
		log.Fatal("迁移失败: ", err)
	}
	fmt.Println("迁移完成")

}
