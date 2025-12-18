package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username string `gorm:"size:32;not null;unique" json:"username"`
	Password string `gorm:"size:128;not null" json:"password"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
