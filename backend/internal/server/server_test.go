package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCorsWithAllowsConfiguredOriginOnly(t *testing.T) {
	handler := corsWith([]string{"https://lingshu.example.com"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	allowed := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://lingshu.example.com")
	handler.ServeHTTP(allowed, req)
	if got := allowed.Header().Get("Access-Control-Allow-Origin"); got != "https://lingshu.example.com" {
		t.Fatalf("allowed origin header=%q", got)
	}

	blocked := httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	handler.ServeHTTP(blocked, req)
	if got := blocked.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("blocked origin header=%q", got)
	}
}

func TestGatewayMaxBodyBytesDefaultsToTwoMiB(t *testing.T) {
	if got := gatewayMaxBodyBytes(0); got != 2*1024*1024 {
		t.Fatalf("limit=%d want %d", got, 2*1024*1024)
	}
}
