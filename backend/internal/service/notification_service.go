package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"lingshu/backend/internal/repository"
)

type NotificationService struct {
	settings repository.SettingsRepository
	client   *http.Client
}

type AlertNotification struct {
	RuleKey    string `json:"rule_key"`
	Severity   string `json:"severity"`
	Title      string `json:"title"`
	Message    string `json:"message"`
	TargetType string `json:"target_type"`
	TargetID   string `json:"target_id"`
}

func NewNotificationService(settings repository.SettingsRepository) NotificationService {
	return NotificationService{
		settings: settings,
		client:   &http.Client{Timeout: 5 * time.Second},
	}
}

func (s NotificationService) SendAlert(ctx context.Context, alert AlertNotification) error {
	settings, err := s.settings.GetMap(ctx,
		"alert_email_recipients", "alert_webhook_url", "alert_webhook_provider",
		"smtp_host", "smtp_port", "smtp_user", "smtp_pass", "smtp_from", "smtp_tls",
	)
	if err != nil {
		return err
	}
	var errs []error
	if recipients := splitRecipients(settings["alert_email_recipients"]); len(recipients) > 0 {
		if err := s.sendEmail(settings, recipients, alert); err != nil {
			errs = append(errs, err)
		}
	}
	if webhook := strings.TrimSpace(settings["alert_webhook_url"]); webhook != "" {
		provider := strings.TrimSpace(settings["alert_webhook_provider"])
		if err := s.sendWebhook(ctx, webhook, provider, alert); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (s NotificationService) sendEmail(settings map[string]string, recipients []string, alert AlertNotification) error {
	port, _ := strconv.Atoi(firstNonEmpty(settings["smtp_port"], "587"))
	from := strings.TrimSpace(firstNonEmpty(settings["smtp_from"], settings["smtp_user"]))
	host := strings.TrimSpace(settings["smtp_host"])
	user := strings.TrimSpace(settings["smtp_user"])
	pass := settings["smtp_pass"]
	if host == "" || port <= 0 || user == "" || pass == "" || from == "" {
		return nil
	}
	subject := "[LingShu Alert] " + alert.Title
	body := fmt.Sprintf("%s\n\nSeverity: %s\nRule: %s\nTarget: %s/%s\n", alert.Message, alert.Severity, alert.RuleKey, alert.TargetType, alert.TargetID)
	message := "From: " + from + "\r\n" +
		"To: " + strings.Join(recipients, ",") + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n\r\n" +
		body
	addr := host + ":" + strconv.Itoa(port)
	auth := smtp.PlainAuth("", user, pass, host)
	if firstNonEmpty(settings["smtp_tls"], "true") == "true" {
		return sendMailTLS(addr, host, auth, from, recipients, []byte(message))
	}
	return smtp.SendMail(addr, auth, from, recipients, []byte(message))
}

func (s NotificationService) sendWebhook(ctx context.Context, webhook, provider string, alert AlertNotification) error {
	payload, err := json.Marshal(formatAlertWebhookPayload(provider, alert))
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhook, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned %s", resp.Status)
	}
	return nil
}

func formatAlertWebhookPayload(provider string, alert AlertNotification) any {
	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider == "" {
		provider = "generic"
	}
	text := fmt.Sprintf("**%s**\n\n%s\n\nSeverity: %s\nRule: %s\nTarget: %s/%s", alert.Title, alert.Message, alert.Severity, alert.RuleKey, alert.TargetType, alert.TargetID)
	switch provider {
	case "wechat", "wecom", "qyweixin":
		return map[string]any{
			"msgtype": "markdown",
			"markdown": map[string]string{
				"content": strings.ReplaceAll(text, "**", ""),
			},
		}
	case "feishu", "lark":
		return map[string]any{
			"msg_type": "interactive",
			"card": map[string]any{
				"header": map[string]any{
					"title":    map[string]string{"tag": "plain_text", "content": alert.Title},
					"template": severityTemplate(alert.Severity),
				},
				"elements": []map[string]any{
					{"tag": "div", "text": map[string]string{"tag": "lark_md", "content": text}},
				},
			},
		}
	case "dingtalk":
		return map[string]any{
			"msgtype": "markdown",
			"markdown": map[string]string{
				"title": alert.Title,
				"text":  text,
			},
		}
	case "discord":
		return map[string]any{
			"embeds": []map[string]any{{
				"title":       alert.Title,
				"description": alert.Message,
				"color":       severityColor(alert.Severity),
				"fields": []map[string]string{
					{"name": "Severity", "value": alert.Severity, "inline": "true"},
					{"name": "Rule", "value": alert.RuleKey, "inline": "true"},
					{"name": "Target", "value": alert.TargetType + "/" + alert.TargetID, "inline": "false"},
				},
			}},
		}
	default:
		return alert
	}
}

func severityTemplate(severity string) string {
	switch strings.ToLower(severity) {
	case "critical":
		return "red"
	case "warning":
		return "orange"
	default:
		return "blue"
	}
}

func severityColor(severity string) int {
	switch strings.ToLower(severity) {
	case "critical":
		return 0xD92D20
	case "warning":
		return 0xDC6803
	default:
		return 0x1570EF
	}
}

func splitRecipients(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
