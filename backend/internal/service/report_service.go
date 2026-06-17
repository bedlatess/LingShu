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

func (s ReportService) UserLedger(ctx context.Context, userID string) ([]repository.LedgerRecord, error) {
	return s.reports.UserLedger(ctx, userID)
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

func (s ReportService) AdminLedger(ctx context.Context) ([]repository.LedgerRecord, error) {
	return s.reports.AdminLedger(ctx)
}
