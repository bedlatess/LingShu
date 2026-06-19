package service

import (
	"context"

	"lingshu/backend/internal/repository"
)

type ReportService struct {
	reports repository.ReportRepository
}

func NewReportService(reports repository.ReportRepository) ReportService {
	return ReportService{reports: reports}
}

func (s ReportService) UserLogs(ctx context.Context, userID string) ([]repository.GatewayLog, error) {
	return s.reports.UserLogs(ctx, userID)
}

func (s ReportService) UserLogsPaged(ctx context.Context, userID string, page, limit int) ([]repository.GatewayLog, int, error) {
	return s.reports.UserLogsPaged(ctx, userID, limit, (page-1)*limit)
}

func (s ReportService) UserLogsFilteredPaged(ctx context.Context, userID string, filter repository.UserLogFilter, page, limit int) ([]repository.GatewayLog, int, error) {
	return s.reports.UserLogsFilteredPaged(ctx, userID, filter, limit, (page-1)*limit)
}

func (s ReportService) UserLedger(ctx context.Context, userID string) ([]repository.LedgerRecord, error) {
	return s.reports.UserLedger(ctx, userID)
}

func (s ReportService) UserLedgerPaged(ctx context.Context, userID string, page, limit int) ([]repository.LedgerRecord, int, error) {
	return s.reports.UserLedgerPaged(ctx, userID, limit, (page-1)*limit)
}

func (s ReportService) UserLedgerFilteredPaged(ctx context.Context, userID string, filter repository.UserLedgerFilter, page, limit int) ([]repository.LedgerRecord, int, error) {
	return s.reports.UserLedgerFilteredPaged(ctx, userID, filter, limit, (page-1)*limit)
}

func (s ReportService) UserSummary(ctx context.Context, userID string) (repository.UserSummaryStats, error) {
	return s.reports.UserSummary(ctx, userID)
}

func (s ReportService) DailyStats(ctx context.Context, userID string, days int) ([]repository.DailyStat, error) {
	if days <= 0 {
		days = 7
	}
	return s.reports.DailyStats(ctx, userID, days)
}

func (s ReportService) ModelStats(ctx context.Context, userID string) ([]repository.ModelStat, error) {
	return s.reports.ModelStats(ctx, userID)
}

func (s ReportService) AdminDashboard(ctx context.Context) (repository.AdminDashboard, error) {
	return s.reports.AdminDashboard(ctx)
}

func (s ReportService) AdminLogs(ctx context.Context) ([]repository.GatewayLog, error) {
	return s.reports.AdminLogs(ctx)
}

func (s ReportService) AdminLogsPaged(ctx context.Context, page, limit int) ([]repository.GatewayLog, int, error) {
	return s.reports.AdminLogsPaged(ctx, limit, (page-1)*limit)
}

func (s ReportService) ExportAdminLogs(ctx context.Context, fn func(repository.GatewayLog) error) error {
	return s.reports.ExportAdminLogs(ctx, fn)
}

func (s ReportService) AdminLedger(ctx context.Context) ([]repository.LedgerRecord, error) {
	return s.reports.AdminLedger(ctx)
}

func (s ReportService) ExportAdminLedger(ctx context.Context, fn func(repository.LedgerRecord) error) error {
	return s.reports.ExportAdminLedger(ctx, fn)
}

func (s ReportService) ExportUserLogs(ctx context.Context, userID string, fn func(repository.GatewayLog) error) error {
	return s.reports.ExportUserLogs(ctx, userID, fn)
}

func (s ReportService) AdminLedgerPaged(ctx context.Context, page, limit int) ([]repository.LedgerRecord, int, error) {
	return s.reports.AdminLedgerPaged(ctx, limit, (page-1)*limit)
}

func (s ReportService) ReportDaily(ctx context.Context, from, to string) ([]repository.ReportRow, error) {
	return s.reports.ReportDaily(ctx, from, to)
}

func (s ReportService) ReportByUser(ctx context.Context, from, to string) ([]repository.ReportRow, error) {
	return s.reports.ReportByUser(ctx, from, to)
}

func (s ReportService) ReportByModel(ctx context.Context, from, to string) ([]repository.ReportRow, error) {
	return s.reports.ReportByModel(ctx, from, to)
}

func (s ReportService) ReportByChannel(ctx context.Context, from, to string) ([]repository.ReportRow, error) {
	return s.reports.ReportByChannel(ctx, from, to)
}
