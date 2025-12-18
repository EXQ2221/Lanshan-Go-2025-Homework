package model

import "gorm.io/gorm"

type Todo struct {
	gorm.Model
	Title  string `gorm:"size:100;not null" json:"title"`
	Done   bool   `gorm:"default:false" json:"done"`
	UserID uint   `gorm:"not null" json:"user_id"`
}

type UpdateDataRequest struct {
	Title *string `json:"title"`
	Done  *bool   `json:"done"`
}

type DeleteDataRequest struct {
	Title  string `json:"title"`
	UserID uint   `json:"user_id"`
}
