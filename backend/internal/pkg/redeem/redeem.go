package redeem

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func GenerateCode() (string, error) {
	bytes := make([]byte, 12)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "LS-" + strings.ToUpper(hex.EncodeToString(bytes)), nil
}

func Hash(code string) string {
	normalized := strings.ToUpper(strings.TrimSpace(code))
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:])
}

func Prefix(code string) string {
	code = strings.TrimSpace(code)
	if len(code) <= 8 {
		return code
	}
	return code[:8]
}
