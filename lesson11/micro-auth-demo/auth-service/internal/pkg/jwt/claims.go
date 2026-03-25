package jwt

import (
	"strconv"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID    int64  `json:"user_id"`
	SessionID string `json:"sid"`
	TokenID   string `json:"jti"`
	jwtv5.RegisteredClaims
}

func NewClaims(userID int64, sessionID, tokenID string, issuedAt, expiresAt time.Time) Claims {
	return Claims{
		UserID:    userID,
		SessionID: sessionID,
		TokenID:   tokenID,
		RegisteredClaims: jwtv5.RegisteredClaims{
			Subject:   strconv.FormatInt(userID, 10),
			ID:        tokenID,
			IssuedAt:  jwtv5.NewNumericDate(issuedAt),
			ExpiresAt: jwtv5.NewNumericDate(expiresAt),
		},
	}
}
