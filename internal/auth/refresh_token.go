package auth

import (
	"crypto/rand"
	"encoding/hex"
)

const tokenByteLentgh = 32

func MakeRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
