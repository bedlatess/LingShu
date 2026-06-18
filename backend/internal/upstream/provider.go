package upstream

import (
	"context"
	"net/http"
	"strings"
)

type Provider interface {
	ForwardChat(ctx context.Context, baseURL, apiKey string, timeoutSeconds int, rawBody []byte, upstreamModelName string) (ChatResponse, error)
	OpenChatStream(ctx context.Context, baseURL, apiKey string, timeoutSeconds int, rawBody []byte, upstreamModelName string) (*http.Response, error)
}

func ProviderForType(providerType string) Provider {
	switch strings.ToLower(strings.TrimSpace(providerType)) {
	case "anthropic", "claude":
		// claude is a legacy alias kept for existing data; new channels use anthropic.
		return AnthropicAdapter{}
	default:
		return OpenAIAdapter{}
	}
}
