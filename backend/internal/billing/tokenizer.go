package billing

import (
	"bufio"
	"encoding/json"
	"strings"
	"sync"
	"unicode/utf8"

	tiktoken "github.com/pkoukk/tiktoken-go"
)

var tokenizer = struct {
	sync.Once
	encoding *tiktoken.Tiktoken
	err      error
}{}

func EstimateTokens(text string) int64 {
	if text == "" {
		return 0
	}
	tokenizer.Once.Do(func() {
		tokenizer.encoding, tokenizer.err = tiktoken.GetEncoding("cl100k_base")
	})
	if tokenizer.err == nil && tokenizer.encoding != nil {
		return int64(len(tokenizer.encoding.Encode(text, nil, nil)))
	}
	return fallbackEstimateTokens(text)
}

func EstimateStreamTokens(raw string) int64 {
	return EstimateTokens(ExtractSSEText(raw))
}

func ExtractSSEText(raw string) string {
	if raw == "" {
		return ""
	}
	var out strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(raw))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "" || data == "[DONE]" {
			continue
		}
		if text := extractSSEDeltaText(data); text != "" {
			out.WriteString(text)
		}
	}
	if out.Len() == 0 {
		return raw
	}
	return out.String()
}

func extractSSEDeltaText(data string) string {
	var payload struct {
		Choices []struct {
			Delta struct {
				Content any `json:"content"`
			} `json:"delta"`
			Message struct {
				Content any `json:"content"`
			} `json:"message"`
			Text string `json:"text"`
		} `json:"choices"`
	}
	if err := json.Unmarshal([]byte(data), &payload); err != nil {
		return ""
	}
	var out strings.Builder
	for _, choice := range payload.Choices {
		writeContent(&out, choice.Delta.Content)
		writeContent(&out, choice.Message.Content)
		if choice.Text != "" {
			out.WriteString(choice.Text)
		}
	}
	return out.String()
}

func writeContent(out *strings.Builder, content any) {
	switch value := content.(type) {
	case string:
		out.WriteString(value)
	case []any:
		for _, part := range value {
			object, ok := part.(map[string]any)
			if !ok {
				continue
			}
			if text, ok := object["text"].(string); ok {
				out.WriteString(text)
			}
		}
	}
}

func fallbackEstimateTokens(text string) int64 {
	runes := int64(utf8.RuneCountInString(text))
	if runes == 0 {
		return 0
	}
	return (runes + 2) / 3
}
