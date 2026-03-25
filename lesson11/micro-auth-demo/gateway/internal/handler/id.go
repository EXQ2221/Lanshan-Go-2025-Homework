package handler

import (
	"crypto/rand"
	"encoding/hex"
)

func newID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "web-client"
	}
	return hex.EncodeToString(buf)
}
