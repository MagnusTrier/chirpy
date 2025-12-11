package auth

import (
	"crypto/rand"
	"encoding/hex"
)

func MakeRefreshToken() (string, error) {
	randomData := make([]byte, 32)
	rand.Read(randomData)
	token := hex.EncodeToString(randomData)
	return token, nil
}
