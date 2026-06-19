package service

import "testing"

func TestFormatAlertWebhookPayloadProviders(t *testing.T) {
	alert := AlertNotification{
		RuleKey:    "channel_consecutive_failures",
		Severity:   "critical",
		Title:      "渠道连续失败",
		Message:    "渠道 A 已达到自动禁用阈值",
		TargetType: "channel",
		TargetID:   "00000000-0000-0000-0000-000000000001",
	}
	tests := []struct {
		name     string
		provider string
		wantKey  string
	}{
		{name: "wechat", provider: "wechat", wantKey: "msgtype"},
		{name: "feishu", provider: "feishu", wantKey: "msg_type"},
		{name: "dingtalk", provider: "dingtalk", wantKey: "markdown"},
		{name: "discord", provider: "discord", wantKey: "embeds"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload, ok := formatAlertWebhookPayload(tt.provider, alert).(map[string]any)
			if !ok {
				t.Fatalf("payload is %T, want map[string]any", payload)
			}
			if _, ok := payload[tt.wantKey]; !ok {
				t.Fatalf("payload missing key %q: %#v", tt.wantKey, payload)
			}
		})
	}
}
