package apikey

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func Generate(prefix string) (string, error) {
	bytes := make([]byte, 24)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return prefix + hex.EncodeToString(bytes), nil
}

func Hash(key string) string {
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:])
}

func Prefix(key string) string {
	if len(key) <= 12 {
		return key
	}
	return key[:12]
}

func Mask(key string) string {
	if len(key) <= 16 {
		return key
	}
	return key[:12] + "..." + key[len(key)-4:]
}

func MaskStored(prefix, suffix string) string {
	return strings.TrimSpace(prefix) + "..." + suffix
}
