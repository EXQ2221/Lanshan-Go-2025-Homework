package main

import (
	"fmt"
	"lesson7/api"
	"lesson7/dao"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	dao.InitDB()
	api.InitRouter()
	dsn := "root:123456@tcp(127.0.0.1:3306)/todolist?charset=utf8mb4&parseTime=True&loc=Local"

	_, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("connect error:", err)
	}

	fmt.Println("connect success")

}
