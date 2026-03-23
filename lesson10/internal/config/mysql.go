package config

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	name := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		user, pass, host, port, name,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("can not connect mysql", err)
	}

	DB = db

}

func CleanupPolymorphicTargetConstraints() error {
	constraints := []struct {
		table string
		name  string
	}{
		{table: "comments", name: "fk_posts_comments"},
		{table: "reactions", name: "fk_posts_reactions"},
		{table: "favorites", name: "fk_posts_favorites"},
		{table: "activities", name: "fk_posts_activities"},
	}

	for _, constraint := range constraints {
		if err := dropForeignKeyIfExists(DB, constraint.table, constraint.name); err != nil {
			return err
		}
	}

	return nil
}

func dropForeignKeyIfExists(db *gorm.DB, tableName, constraintName string) error {
	var count int64
	err := db.Raw(`
		SELECT COUNT(*)
		FROM information_schema.TABLE_CONSTRAINTS
		WHERE CONSTRAINT_SCHEMA = DATABASE()
		  AND TABLE_NAME = ?
		  AND CONSTRAINT_NAME = ?
		  AND CONSTRAINT_TYPE = 'FOREIGN KEY'
	`, tableName, constraintName).Scan(&count).Error
	if err != nil {
		return err
	}

	if count == 0 {
		return nil
	}

	return db.Exec(fmt.Sprintf("ALTER TABLE `%s` DROP FOREIGN KEY `%s`", tableName, constraintName)).Error
}
