package dao

import (
	"lesson6/model"
	"time"
)

var database = map[string]string{
	"initialize": "123456",
}
var refreshTokens = map[string]model.RefreshToken{}

func AddUser(username string, password string) {
	database[username] = password
}

func CheckUser(username, password string) bool {
	pwd, ok := database[username]
	if !ok {
		return false
	}
	return pwd == password
}

func GetPassword(username string) (string, bool) {
	pwd, ok := database[username]
	return pwd, ok
}

func UserExist(username string) bool {
	_, ok := database[username]
	return ok
}

func ValidateRefreshToken(tokenString string) (string, bool) {
	record, ok := refreshTokens[tokenString]
	if !ok {
		return "", false
	}

	if time.Now().After(record.Exp) {
		delete(refreshTokens, tokenString)
		SaveRefreshToken()
		return "", false
	}
	username := record.Username
	delete(refreshTokens, tokenString)
	SaveRefreshToken()
	return username, true
}

func AddRefreshToken(tokenString, username string) {
	refreshTokens[tokenString] = model.RefreshToken{
		Username: username,
		Exp:      time.Now().Add(7 * 24 * time.Hour), // 7天有效期
	}
	SaveRefreshToken()
}
