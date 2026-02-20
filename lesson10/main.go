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
	err := db.AutoMigrate(&core.User{}) // 注意是 models.User，不是 core.User
	if err != nil {
		log.Fatal("迁移失败: ", err)
	}
	fmt.Println("迁移完成")
}
