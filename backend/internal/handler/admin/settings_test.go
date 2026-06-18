package admin

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseAuditLogFilterDateRange(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/admin/audit-logs?actor_id=abc&action=admin.user.update&target_type=user&from=2026-06-01&to=2026-06-18", nil)

	filter, err := parseAuditLogFilter(req)
	if err != nil {
		t.Fatalf("parseAuditLogFilter returned error: %v", err)
	}
	if filter.ActorID != "abc" {
		t.Fatalf("actor id = %q, want %q", filter.ActorID, "abc")
	}
	if filter.Action != "admin.user.update" {
		t.Fatalf("action = %q, want %q", filter.Action, "admin.user.update")
	}
	if filter.TargetType != "user" {
		t.Fatalf("target type = %q, want %q", filter.TargetType, "user")
	}
	if filter.From == nil || filter.From.Format("2006-01-02") != "2026-06-01" {
		t.Fatalf("from = %v, want 2026-06-01", filter.From)
	}
	if filter.To == nil || filter.To.Format("2006-01-02") != "2026-06-18" {
		t.Fatalf("to = %v, want 2026-06-18", filter.To)
	}
}
