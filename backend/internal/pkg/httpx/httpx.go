package httpx

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func JSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, ErrorResponse{Error: message})
}

func ErrorJSON(w http.ResponseWriter, status int, errType, message, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]any{
			"message": message,
			"type":    errType,
			"code":    code,
		},
	})
}

func Decode(r *http.Request, target any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}

func DeviceID(r *http.Request) string {
	value := strings.TrimSpace(r.Header.Get("X-Device-Id"))
	if len(value) > 128 {
		value = value[:128]
	}
	return value
}

func VerifiedDeviceID(r *http.Request, secret string) string {
	deviceID := DeviceID(r)
	secret = strings.TrimSpace(secret)
	signature := strings.TrimSpace(r.Header.Get("X-Device-Sign"))
	if deviceID == "" || secret == "" || signature == "" {
		return ""
	}
	expected := DeviceSignature(deviceID, r.UserAgent(), secret)
	if !hmac.Equal([]byte(strings.ToLower(signature)), []byte(expected)) {
		return ""
	}
	return deviceID
}

func DeviceSignature(deviceID, userAgent, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(deviceID + userAgent))
	return hex.EncodeToString(mac.Sum(nil))
}
