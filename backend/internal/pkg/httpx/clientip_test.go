package httpx

import (
	"context"
	"net/http/httptest"
	"testing"
)

type testSettings map[string]string

func (s testSettings) GetMap(ctx context.Context, keys ...string) (map[string]string, error) {
	out := map[string]string{}
	for _, key := range keys {
		out[key] = s[key]
	}
	return out, nil
}

func TestClientIPDoesNotTrustProxyHeadersByDefault(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.9:1234"
	req.Header.Set("X-Forwarded-For", "203.0.113.10")

	got := ClientIP(req, nil)
	if got != "10.0.0.9" {
		t.Fatalf("ClientIP = %q, want remote addr", got)
	}
}

func TestClientIPTrustsProxyHeadersWhenEnabled(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.9:1234"
	req.Header.Set("X-Forwarded-For", "203.0.113.10, 198.51.100.3")

	got := ClientIP(req, testSettings{
		"trusted_proxy_enabled": "true",
		"trusted_proxy_hops":    "2",
	})
	if got != "203.0.113.10" {
		t.Fatalf("ClientIP = %q, want first forwarded hop", got)
	}
}

func TestVerifiedDeviceID(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("User-Agent", "test-agent")
	req.Header.Set("X-Device-Id", "dev_test_123456")
	req.Header.Set("X-Device-Sign", DeviceSignature("dev_test_123456", "test-agent", "secret"))

	got := VerifiedDeviceID(req, "secret")
	if got != "dev_test_123456" {
		t.Fatalf("VerifiedDeviceID = %q, want device id", got)
	}
}

func TestVerifiedDeviceIDRejectsBadSignature(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("User-Agent", "test-agent")
	req.Header.Set("X-Device-Id", "dev_test_123456")
	req.Header.Set("X-Device-Sign", "bad")

	got := VerifiedDeviceID(req, "secret")
	if got != "" {
		t.Fatalf("VerifiedDeviceID = %q, want empty for bad signature", got)
	}
}
