package upstream

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
)

// ProbeOpenAI strictly probes an OpenAI-compatible upstream with GET <base>/models.
// It never falls back to presets.
func ProbeOpenAI(ctx context.Context, baseURL, apiKey string) ([]string, int, error) {
	url := strings.TrimRight(strings.TrimSpace(baseURL), "/") + "/models"
	return probeModels(ctx, http.MethodGet, url, func(req *http.Request) {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	})
}

// ProbeAnthropic strictly probes Anthropic with GET <base>/v1/models.
// It never falls back to hard-coded Claude models.
func ProbeAnthropic(ctx context.Context, baseURL, apiKey string) ([]string, int, error) {
	return probeModels(ctx, http.MethodGet, anthropicURL(baseURL, "/models"), func(req *http.Request) {
		req.Header.Set("x-api-key", apiKey)
		req.Header.Set("anthropic-version", "2023-06-01")
	})
}

func probeModels(ctx context.Context, method, url string, auth func(*http.Request)) ([]string, int, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, 0, err
	}
	auth(req)
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, resp.StatusCode, errors.New(strings.TrimSpace(string(body)))
	}
	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	if !strings.Contains(contentType, "application/json") {
		return nil, resp.StatusCode, &ProviderContentTypeError{StatusCode: resp.StatusCode, ContentType: resp.Header.Get("Content-Type"), Body: string(body)}
	}
	var parsed struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, resp.StatusCode, errors.New("上游返回非 JSON，请检查 base_url 是否正确")
	}
	ids := make([]string, 0, len(parsed.Data))
	for _, item := range parsed.Data {
		if id := strings.TrimSpace(item.ID); id != "" {
			ids = append(ids, id)
		}
	}
	return ids, resp.StatusCode, nil
}

type DetectResult struct {
	Format         string   `json:"format"`
	NormalizedBase string   `json:"normalized_base_url"`
	SampleModels   []string `json:"sample_models"`
}

func DetectProtocol(ctx context.Context, baseURL, apiKey string) (DetectResult, error) {
	trimmed := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	lower := strings.ToLower(trimmed)
	if trimmed == "" || strings.TrimSpace(apiKey) == "" {
		return DetectResult{}, errors.New("base_url 和 api_key 不能为空")
	}

	if strings.Contains(lower, "anthropic") || strings.Contains(lower, "claude") {
		if models, _, err := ProbeAnthropic(ctx, trimmed, apiKey); err == nil {
			return DetectResult{Format: "anthropic", NormalizedBase: trimmed, SampleModels: models}, nil
		}
	}
	if models, _, err := ProbeOpenAI(ctx, trimmed, apiKey); err == nil {
		return DetectResult{Format: "openai", NormalizedBase: trimmed, SampleModels: models}, nil
	}
	if !strings.Contains(lower, "/v") {
		withV1 := trimmed + "/v1"
		if models, _, err := ProbeOpenAI(ctx, withV1, apiKey); err == nil {
			return DetectResult{Format: "openai", NormalizedBase: withV1, SampleModels: models}, nil
		}
	}
	if models, _, err := ProbeAnthropic(ctx, trimmed, apiKey); err == nil {
		return DetectResult{Format: "anthropic", NormalizedBase: trimmed, SampleModels: models}, nil
	}
	return DetectResult{}, errors.New("无法识别上游格式：OpenAI /models 与 Anthropic /v1/models 均探测失败，请检查 base_url 与密钥")
}
