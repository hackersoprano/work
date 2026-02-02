package services

import (
	"crypto/sha256"
	"encoding/hex"
)

func HashPassword(password string) string {
	hasher := sha256.New()
	hasher.Write([]byte(password))
	return hex.EncodeToString(hasher.Sum(nil)) //возврат значения и преобразование из бит в 16 с.ч.
}
