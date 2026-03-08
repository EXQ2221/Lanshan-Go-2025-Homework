package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

func NewSID() (string, error) {
	b := make([]byte, 16) // 128-bit
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil // 32位十六进制字符串
}

func HashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
