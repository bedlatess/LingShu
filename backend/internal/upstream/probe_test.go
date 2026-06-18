package upstream

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDetectProtocolOpenAI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/models" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("unexpected auth: %s", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"id":"gpt-x"}]}`))
	}))
	defer server.Close()

	result, err := DetectProtocol(context.Background(), server.URL, "test-key")
	if err != nil {
		t.Fatalf("DetectProtocol failed: %v", err)
	}
	if result.Format != "openai" || result.NormalizedBase != server.URL || len(result.SampleModels) != 1 || result.SampleModels[0] != "gpt-x" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestDetectProtocolAddsV1ForOpenAI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/models" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"id":"gpt-v1"}]}`))
	}))
	defer server.Close()

	result, err := DetectProtocol(context.Background(), server.URL, "test-key")
	if err != nil {
		t.Fatalf("DetectProtocol failed: %v", err)
	}
	if result.Format != "openai" || !strings.HasSuffix(result.NormalizedBase, "/v1") {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestDetectProtocolAnthropic(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/anthropic/v1/models" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("x-api-key"); got != "test-key" {
			t.Fatalf("unexpected api key: %s", got)
		}
		if got := r.Header.Get("anthropic-version"); got != "2023-06-01" {
			t.Fatalf("unexpected anthropic version: %s", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"id":"claude-x"}]}`))
	}))
	defer server.Close()

	result, err := DetectProtocol(context.Background(), server.URL+"/anthropic", "test-key")
	if err != nil {
		t.Fatalf("DetectProtocol failed: %v", err)
	}
	if result.Format != "anthropic" || len(result.SampleModels) != 1 || result.SampleModels[0] != "claude-x" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestDetectProtocolWithAnthropicHintPrefersAnthropicProbe(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/models":
			if got := r.Header.Get("x-api-key"); got != "test-key" {
				t.Fatalf("unexpected api key: %s", got)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"data":[{"id":"claude-hint"}]}`))
		case "/models":
			t.Fatalf("anthropic hint should not probe OpenAI /models first")
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	result, err := DetectProtocolWithHint(context.Background(), server.URL, "test-key", "anthropic")
	if err != nil {
		t.Fatalf("DetectProtocolWithHint failed: %v", err)
	}
	if result.Format != "anthropic" || result.NormalizedBase != server.URL || result.ProbeURL != server.URL+"/v1/models" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestProbeAnthropicStrictNoPresetFallback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":{"message":"bad key"}}`, http.StatusUnauthorized)
	}))
	defer server.Close()

	models, status, err := ProbeAnthropic(context.Background(), server.URL, "bad-key")
	if err == nil {
		t.Fatal("expected ProbeAnthropic to fail")
	}
	if status != http.StatusUnauthorized {
		t.Fatalf("unexpected status: %d", status)
	}
	if len(models) != 0 {
		t.Fatalf("strict probe must not return preset models: %+v", models)
	}
}
