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
		"alert_email_recipients", "alert_webhook_url",
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
		if err := s.sendWebhook(ctx, webhook, alert); err != nil {
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

func (s NotificationService) sendWebhook(ctx context.Context, webhook string, alert AlertNotification) error {
	payload, err := json.Marshal(alert)
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
