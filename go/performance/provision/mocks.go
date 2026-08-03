package provision

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func GenerateSignature(ts, method, fullURL, body, secret string) string {
	msg := ts + "|" + method + "|" + fullURL + "|" + body
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(msg))
	return strings.Trim(hex.EncodeToString(h.Sum(nil)), "|")
}
