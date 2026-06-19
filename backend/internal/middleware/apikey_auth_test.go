package middleware

import "testing"

func TestEndpointAllowed(t *testing.T) {
	if !endpointAllowed("/v1/chat/completions", nil) {
		t.Fatalf("empty whitelist should allow all endpoints")
	}
	if !endpointAllowed("/v1/chat/completions", []string{"/v1/chat/completions"}) {
		t.Fatalf("matching endpoint should be allowed")
	}
	if !endpointAllowed("/v1/chat/completions/", []string{"/v1/chat/completions"}) {
		t.Fatalf("trailing slash should be normalized")
	}
	if endpointAllowed("/v1/embeddings", []string{"/v1/chat/completions"}) {
		t.Fatalf("non-whitelisted endpoint should be denied")
	}
	if !endpointAllowed("/messages", []string{"/v1/messages"}) {
		t.Fatalf("legacy /v1/messages whitelist should allow /messages")
	}
	if !endpointAllowed("/v1/messages", []string{"/messages"}) {
		t.Fatalf("/messages whitelist should allow legacy /v1/messages")
	}
}
