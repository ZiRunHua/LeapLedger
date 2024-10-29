package util

import (
	"crypto/sha256"
	"encoding/hex"
)

func ClientPasswordHash(email, password string) string {
	hash := sha256.Sum256([]byte(email + password))
	return hex.EncodeToString(hash[:])
}
