package main

import (
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
}
