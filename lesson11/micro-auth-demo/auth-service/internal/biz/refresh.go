package biz

import "example.com/micro-auth-demo/auth-service/internal/pkg/token"

func NewRefreshToken() (string, error) {
	return token.Generate(32)
}
