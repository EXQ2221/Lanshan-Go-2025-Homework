package dao

import (
	"lesson8/model"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	dsn := "root:123456@tcp(127.0.0.1:3306)/todolist?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("can not connect dao", err)
	}

	err = db.AutoMigrate(&model.User{}, &model.Todo{})
	if err != nil {
		log.Fatal("auto migrate error")
	}

	DB = db

}
