package model

import "time"

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshToken struct {
	Username string    `json:"username"`
	Exp      time.Time `json:"exp"`
}
