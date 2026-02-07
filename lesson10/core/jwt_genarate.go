package core

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte(os.Getenv("JWT_SECRET"))

func GenerateToken(username string, userID uint, TokenVersion int, role Role) (string, error) {
	claims := jwt.MapClaims{
		"username":      username,
		"user_id":       userID,
		"token_version": TokenVersion,
		"role":          role,
		"exp":           time.Now().Add(time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(secretKey)
}

func ValidateToken(token string) (*jwt.Token, error) {
	return jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

}
