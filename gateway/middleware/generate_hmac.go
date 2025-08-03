package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func generateHmacHash(secret string, data string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	hash := h.Sum(nil)

	hexHash := hex.EncodeToString(hash)

	return hexHash
}
