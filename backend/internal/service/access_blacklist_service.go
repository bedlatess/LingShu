package service

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"lingshu/backend/internal/repository"
)

var ErrAccessBlacklisted = errors.New("access denied by security policy")

type AccessBlacklistService struct {
	blacklist repository.AccessBlacklistRepository
	settings  repository.SettingsRepository
	audits    repository.AuditRepository
	redis     *redis.Client
}

type AccessSubject struct {
	Scope    string
	IP       string
	DeviceID string
}

type CreateAccessBlacklistRequest struct {
	Kind      string     `json:"kind"`
	Value     string     `json:"value"`
	Scope     string     `json:"scope"`
	Reason    string     `json:"reason"`
	Permanent bool       `json:"permanent"`
	ExpiresAt *time.Time `json:"expires_at"`
}

type AccessBlacklistMatch struct {
	Blocked bool                            `json:"blocked"`
	Entry   repository.AccessBlacklistEntry `json:"entry,omitempty"`
}

func NewAccessBlacklistService(blacklist repository.AccessBlacklistRepository, settings repository.SettingsRepository, audits repository.AuditRepository, redisClient *redis.Client) AccessBlacklistService {
	return AccessBlacklistService{blacklist: blacklist, settings: settings, audits: audits, redis: redisClient}
}

func (s AccessBlacklistService) ListPaged(ctx context.Context, filter repository.AccessBlacklistFilter, page, limit int) ([]repository.AccessBlacklistEntry, int, error) {
	return s.blacklist.ListPaged(ctx, filter, limit, (page-1)*limit)
}

func (s AccessBlacklistService) CreateManual(ctx context.Context, actorID string, input CreateAccessBlacklistRequest, ip, userAgent string) (repository.AccessBlacklistEntry, error) {
	kind, value, scope, err := normalizeBlacklistInput(input.Kind, input.Value, input.Scope)
	if err != nil {
		return repository.AccessBlacklistEntry{}, err
	}
	reason := strings.TrimSpace(input.Reason)
	if reason == "" {
		return repository.AccessBlacklistEntry{}, errors.New("reason is required")
	}
	expiresAt := input.ExpiresAt
	if !input.Permanent && expiresAt == nil {
		defaultExpiry := time.Now().UTC().AddDate(0, 0, s.autoTTLDays(ctx))
		expiresAt = &defaultExpiry
	}
	item, err := s.blacklist.Create(ctx, repository.CreateAccessBlacklistInput{
		Kind:      kind,
		Value:     value,
		Scope:     scope,
		Reason:    reason,
		Source:    "manual",
		CreatedBy: actorID,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return repository.AccessBlacklistEntry{}, err
	}
	s.invalidate(ctx)
	_ = s.audits.Write(ctx, repository.AuditEntry{
		ActorID:    actorID,
		Action:     "admin.access_blacklist.create",
		TargetType: "access_blacklist",
		TargetID:   item.ID,
		After:      item,
		IP:         ip,
		UserAgent:  userAgent,
	})
	return item, nil
}

func (s AccessBlacklistService) Release(ctx context.Context, actorID, id, ip, userAgent string) (repository.AccessBlacklistEntry, error) {
	item, err := s.blacklist.Release(ctx, id, actorID)
	if err != nil {
		return repository.AccessBlacklistEntry{}, err
	}
	s.invalidate(ctx)
	_ = s.audits.Write(ctx, repository.AuditEntry{
		ActorID:    actorID,
		Action:     "admin.access_blacklist.release",
		TargetType: "access_blacklist",
		TargetID:   item.ID,
		After:      item,
		IP:         ip,
		UserAgent:  userAgent,
	})
	return item, nil
}

func (s AccessBlacklistService) Check(ctx context.Context, subject AccessSubject) (AccessBlacklistMatch, error) {
	subject.Scope = normalizeScope(subject.Scope)
	subject.IP = strings.TrimSpace(subject.IP)
	subject.DeviceID = strings.TrimSpace(subject.DeviceID)
	if subject.IP == "" && subject.DeviceID == "" {
		return AccessBlacklistMatch{}, nil
	}
	if cached, ok := s.cachedMatch(ctx, subject); ok {
		return cached, nil
	}
	item, blocked, err := s.blacklist.Matches(ctx, subject.Scope, subject.IP, subject.DeviceID)
	if err != nil {
		return AccessBlacklistMatch{}, err
	}
	match := AccessBlacklistMatch{Blocked: blocked, Entry: item}
	s.cacheMatch(ctx, subject, match)
	return match, nil
}

func (s AccessBlacklistService) DeviceSecret(ctx context.Context) string {
	values, err := s.settings.GetMap(ctx, "device_secret_key")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(values["device_secret_key"])
}

func (s AccessBlacklistService) EnsureAllowed(ctx context.Context, subject AccessSubject) error {
	match, err := s.Check(ctx, subject)
	if err != nil {
		return err
	}
	if match.Blocked {
		return ErrAccessBlacklisted
	}
	return nil
}

func (s AccessBlacklistService) RecordLoginFailure(ctx context.Context, ip, deviceID string) {
	if !s.autoEnabled(ctx) {
		return
	}
	threshold := s.loginFailThreshold(ctx)
	for _, subject := range []struct {
		kind  string
		value string
	}{
		{kind: "ip", value: strings.TrimSpace(ip)},
		{kind: "device", value: strings.TrimSpace(deviceID)},
	} {
		if subject.value == "" {
			continue
		}
		count, err := s.incrFailure(ctx, subject.kind, subject.value)
		if err == nil && count >= threshold {
			s.autoBlock(ctx, subject.kind, subject.value)
		}
	}
}

func (s AccessBlacklistService) RecordLoginSuccess(ctx context.Context, ip, deviceID string) {
	for _, subject := range []struct {
		kind  string
		value string
	}{
		{kind: "ip", value: strings.TrimSpace(ip)},
		{kind: "device", value: strings.TrimSpace(deviceID)},
	} {
		if subject.value == "" || s.redis == nil {
			continue
		}
		_ = s.redis.Del(ctx, loginFailureKey(subject.kind, subject.value)).Err()
	}
}

func (s AccessBlacklistService) autoBlock(ctx context.Context, kind, value string) {
	kind, value, scope, err := normalizeBlacklistInput(kind, value, "all")
	if err != nil {
		return
	}
	expiresAt := time.Now().UTC().AddDate(0, 0, s.autoTTLDays(ctx))
	item, err := s.blacklist.Create(ctx, repository.CreateAccessBlacklistInput{
		Kind:      kind,
		Value:     value,
		Scope:     scope,
		Reason:    "登录失败次数过多，系统自动拉黑",
		Source:    "auto",
		ExpiresAt: &expiresAt,
	})
	if err == nil {
		s.invalidate(ctx)
		_ = s.audits.Write(ctx, repository.AuditEntry{
			Action:     "security.access_blacklist.auto_create",
			TargetType: "access_blacklist",
			TargetID:   item.ID,
			After:      item,
		})
	}
}

func (s AccessBlacklistService) incrFailure(ctx context.Context, kind, value string) (int64, error) {
	if s.redis == nil {
		return 0, errors.New("redis unavailable")
	}
	key := loginFailureKey(kind, value)
	count, err := s.redis.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if count == 1 {
		_ = s.redis.Expire(ctx, key, 30*time.Minute).Err()
	}
	return count, nil
}

func (s AccessBlacklistService) cachedMatch(ctx context.Context, subject AccessSubject) (AccessBlacklistMatch, bool) {
	if s.redis == nil {
		return AccessBlacklistMatch{}, false
	}
	payload, err := s.redis.Get(ctx, blacklistCacheKey(subject)).Bytes()
	if err != nil {
		return AccessBlacklistMatch{}, false
	}
	var match AccessBlacklistMatch
	if err := json.Unmarshal(payload, &match); err != nil {
		return AccessBlacklistMatch{}, false
	}
	return match, true
}

func (s AccessBlacklistService) cacheMatch(ctx context.Context, subject AccessSubject, match AccessBlacklistMatch) {
	if s.redis == nil {
		return
	}
	payload, err := json.Marshal(match)
	if err != nil {
		return
	}
	ttl := 30 * time.Second
	if match.Blocked {
		ttl = time.Minute
	}
	_ = s.redis.Set(ctx, blacklistCacheKey(subject), payload, ttl).Err()
}

func (s AccessBlacklistService) invalidate(ctx context.Context) {
	if s.redis == nil {
		return
	}
	iter := s.redis.Scan(ctx, 0, "access_blacklist:match:*", 0).Iterator()
	keys := []string{}
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if len(keys) > 0 {
		_ = s.redis.Del(ctx, keys...).Err()
	}
}

func (s AccessBlacklistService) autoEnabled(ctx context.Context) bool {
	values, err := s.settings.GetMap(ctx, "access_blacklist_auto_enabled")
	if err != nil {
		return true
	}
	return !strings.EqualFold(strings.TrimSpace(values["access_blacklist_auto_enabled"]), "false")
}

func (s AccessBlacklistService) loginFailThreshold(ctx context.Context) int64 {
	values, err := s.settings.GetMap(ctx, "access_blacklist_login_fail_threshold")
	if err != nil {
		return 10
	}
	parsed, err := strconv.Atoi(strings.TrimSpace(values["access_blacklist_login_fail_threshold"]))
	if err != nil || parsed <= 0 {
		return 10
	}
	return int64(parsed)
}

func (s AccessBlacklistService) autoTTLDays(ctx context.Context) int {
	values, err := s.settings.GetMap(ctx, "access_blacklist_auto_ttl_days")
	if err != nil {
		return 7
	}
	parsed, err := strconv.Atoi(strings.TrimSpace(values["access_blacklist_auto_ttl_days"]))
	if err != nil || parsed <= 0 {
		return 7
	}
	return parsed
}

func normalizeBlacklistInput(kind, value, scope string) (string, string, string, error) {
	kind = strings.TrimSpace(strings.ToLower(kind))
	value = strings.TrimSpace(value)
	scope = normalizeScope(scope)
	if kind != "ip" && kind != "cidr" && kind != "device" {
		return "", "", "", errors.New("kind must be ip, cidr or device")
	}
	if scope != "login" && scope != "gateway" && scope != "all" {
		return "", "", "", errors.New("scope must be login, gateway or all")
	}
	if value == "" {
		return "", "", "", errors.New("value is required")
	}
	switch kind {
	case "ip":
		ip := net.ParseIP(value)
		if ip == nil {
			return "", "", "", errors.New("invalid ip address")
		}
		value = ip.String()
	case "cidr":
		_, network, err := net.ParseCIDR(value)
		if err != nil {
			return "", "", "", errors.New("invalid cidr")
		}
		value = network.String()
	case "device":
		if len(value) < 8 || len(value) > 128 {
			return "", "", "", errors.New("device id length must be 8-128")
		}
	}
	return kind, value, scope, nil
}

func normalizeScope(scope string) string {
	scope = strings.TrimSpace(strings.ToLower(scope))
	if scope == "" {
		return "all"
	}
	return scope
}

func blacklistCacheKey(subject AccessSubject) string {
	return "access_blacklist:match:" + subject.Scope + ":" + subject.IP + ":" + subject.DeviceID
}

func loginFailureKey(kind, value string) string {
	return "access_blacklist:login_fail:" + kind + ":" + value
}
