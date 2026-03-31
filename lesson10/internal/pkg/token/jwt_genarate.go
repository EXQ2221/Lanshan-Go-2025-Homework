package token

import (
	"lesson10/internal/model"
	"lesson10/internal/pkg/utils"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AccessClaims struct {
	UserID    uint       `json:"user_id"`
	Username  string     `json:"username"`
	Role      model.Role `json:"role"`
	SessionID string     `json:"sid"`
	TokenID   string     `json:"jti"`
	jwt.RegisteredClaims
}

func jwtSecret() []byte {
	return []byte(strings.TrimSpace(os.Getenv("JWT_SECRET")))
}

func AccessTTL() time.Duration {
	raw := strings.TrimSpace(os.Getenv("JWT_EXPIRE_HOURS"))
	if raw == "" {
		return time.Hour
	}

	hours, err := strconv.Atoi(raw)
	if err != nil || hours <= 0 {
		return time.Hour
	}

	return time.Duration(hours) * time.Hour
}

func RefreshTTL() time.Duration {
	raw := strings.TrimSpace(os.Getenv("REFRESH_TOKEN_EXPIRE_HOURS"))
	if raw == "" {
		return 7 * 24 * time.Hour
	}

	hours, err := strconv.Atoi(raw)
	if err != nil || hours <= 0 {
		return 7 * 24 * time.Hour
	}

	return time.Duration(hours) * time.Hour
}

func GenerateToken(username string, userID uint, role model.Role, sessionID string) (string, string, time.Time, error) {
	tokenID, err := utils.NewToken(16)
	if err != nil {
		return "", "", time.Time{}, err
	}

	now := time.Now()
	expiresAt := now.Add(AccessTTL())
	claims := AccessClaims{
		UserID:    userID,
		Username:  username,
		Role:      role,
		SessionID: sessionID,
		TokenID:   tokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatUint(uint64(userID), 10),
			ID:        tokenID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	signedToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenValue, err := signedToken.SignedString(jwtSecret())
	if err != nil {
		return "", "", time.Time{}, err
	}

	return tokenValue, tokenID, expiresAt, nil
}

func GenerateRefreshToken() (string, error) {
	return utils.NewToken(32)
}

func ValidateToken(rawToken string) (*AccessClaims, error) {
	tokenValue, err := jwt.ParseWithClaims(rawToken, &AccessClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret(), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return nil, err
	}

	claims, ok := tokenValue.Claims.(*AccessClaims)
	if !ok || !tokenValue.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return claims, nil
}
