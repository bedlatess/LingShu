package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type GatewayLog struct {
	RequestID   string    `json:"request_id"`
	ModelID     string    `json:"model_id"`
	Status      string    `json:"status"`
	HTTPStatus  int       `json:"http_status"`
	TotalTokens int       `json:"total_tokens"`
	BaseCost    string    `json:"base_cost"`
	Charge      string    `json:"charge"`
	CreatedAt   time.Time `json:"created_at"`
}

type LedgerRecord struct {
	Type           string    `json:"type"`
	Amount         string    `json:"amount"`
	BalanceBefore  string    `json:"balance_before"`
	BalanceAfter   string    `json:"balance_after"`
	BaseCost       string    `json:"base_cost"`
	RateMultiplier string    `json:"rate_multiplier"`
	Remark         string    `json:"remark"`
	CreatedAt      time.Time `json:"created_at"`
}

type DailyStat struct {
	Day         string `json:"day"`
	Requests    int    `json:"requests"`
	Successes   int    `json:"successes"`
	Failures    int    `json:"failures"`
	TotalTokens int    `json:"total_tokens"`
	BaseCost    string `json:"base_cost"`
	Charge      string `json:"charge"`
	GrossProfit string `json:"gross_profit"`
}

type ModelStat struct {
	ModelID     string `json:"model_id"`
	Requests    int    `json:"requests"`
	TotalTokens int    `json:"total_tokens"`
	BaseCost    string `json:"base_cost"`
	Charge      string `json:"charge"`
}

type AdminDashboard struct {
	TodayRequests int    `json:"today_requests"`
	TodayCharge   string `json:"today_charge"`
	TodayBaseCost string `json:"today_base_cost"`
	GrossProfit   string `json:"gross_profit"`
	ActiveUsers   int    `json:"active_users"`
	BalanceTotal  string `json:"balance_total"`
}

type UserDashboardStats struct {
	TodayRequests int    `json:"today_requests"`
	TodayCharge   string `json:"today_charge"`
	MonthCharge   string `json:"month_charge"`
}

type ReportRepository struct {
	db *pgxpool.Pool
}

func NewReportRepository(db *pgxpool.Pool) ReportRepository {
	return ReportRepository{db: db}
}

func (r ReportRepository) UserLogs(ctx context.Context, userID string) ([]GatewayLog, error) {
	rows, err := r.db.Query(ctx, `
		SELECT request_id, COALESCE(model_id::text,''), status, http_status, total_tokens, base_cost::text, charge::text, created_at
		FROM gateway_requests
		WHERE user_id=$1
		ORDER BY created_at DESC
		LIMIT 100
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanGatewayLogs(rows)
}

func (r ReportRepository) UserLedger(ctx context.Context, userID string) ([]LedgerRecord, error) {
	rows, err := r.db.Query(ctx, `
		SELECT type, amount::text, balance_before::text, balance_after::text,
		       COALESCE(base_cost::text,'0'), COALESCE(rate_multiplier::text,'0'), remark, created_at
		FROM balance_ledger
		WHERE user_id=$1
		ORDER BY created_at DESC
		LIMIT 100
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanLedger(rows)
}

func (r ReportRepository) DailyStats(ctx context.Context, userID string, days int) ([]DailyStat, error) {
	rows, err := r.db.Query(ctx, `
		SELECT to_char(date_trunc('day', created_at), 'YYYY-MM-DD') AS day,
		       count(*)::int,
		       count(*) FILTER (WHERE status='success')::int,
		       count(*) FILTER (WHERE status!='success')::int,
		       COALESCE(sum(total_tokens),0)::int,
		       COALESCE(sum(base_cost),0)::text,
		       COALESCE(sum(charge),0)::text,
		       COALESCE(sum(charge - base_cost),0)::text
		FROM gateway_requests
		WHERE user_id=$1 AND created_at >= now() - ($2::int || ' days')::interval
		GROUP BY 1
		ORDER BY 1
	`, userID, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDaily(rows)
}

func (r ReportRepository) ModelStats(ctx context.Context, userID string) ([]ModelStat, error) {
	rows, err := r.db.Query(ctx, `
		SELECT COALESCE(model_id::text,''), count(*)::int, COALESCE(sum(total_tokens),0)::int,
		       COALESCE(sum(base_cost),0)::text, COALESCE(sum(charge),0)::text
		FROM gateway_requests
		WHERE user_id=$1
		GROUP BY model_id
		ORDER BY sum(charge) DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanModelStats(rows)
}

func (r ReportRepository) AdminDashboard(ctx context.Context) (AdminDashboard, error) {
	var item AdminDashboard
	err := r.db.QueryRow(ctx, `
		SELECT
		  COALESCE((SELECT count(*) FROM gateway_requests WHERE created_at::date = now()::date),0)::int,
		  COALESCE((SELECT sum(charge) FROM gateway_requests WHERE created_at::date = now()::date),0)::text,
		  COALESCE((SELECT sum(base_cost) FROM gateway_requests WHERE created_at::date = now()::date),0)::text,
		  COALESCE((SELECT sum(charge - base_cost) FROM gateway_requests),0)::text,
		  COALESCE((SELECT count(*) FROM users WHERE status='active'),0)::int,
		  COALESCE((SELECT sum(balance) FROM users),0)::text
	`).Scan(&item.TodayRequests, &item.TodayCharge, &item.TodayBaseCost, &item.GrossProfit, &item.ActiveUsers, &item.BalanceTotal)
	return item, err
}

func (r ReportRepository) UserDashboard(ctx context.Context, userID string) (UserDashboardStats, error) {
	var item UserDashboardStats
	err := r.db.QueryRow(ctx, `
		SELECT
		  COALESCE((SELECT count(*) FROM gateway_requests WHERE user_id=$1 AND created_at::date = now()::date),0)::int,
		  COALESCE((SELECT sum(charge) FROM gateway_requests WHERE user_id=$1 AND created_at::date = now()::date),0)::text,
		  COALESCE((SELECT sum(charge) FROM gateway_requests WHERE user_id=$1 AND date_trunc('month', created_at)=date_trunc('month', now())),0)::text
	`, userID).Scan(&item.TodayRequests, &item.TodayCharge, &item.MonthCharge)
	return item, err
}

func (r ReportRepository) AdminLogs(ctx context.Context) ([]GatewayLog, error) {
	rows, err := r.db.Query(ctx, `
		SELECT request_id, COALESCE(model_id::text,''), status, http_status, total_tokens, base_cost::text, charge::text, created_at
		FROM gateway_requests
		ORDER BY created_at DESC
		LIMIT 200
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanGatewayLogs(rows)
}

func (r ReportRepository) AdminLedger(ctx context.Context) ([]LedgerRecord, error) {
	rows, err := r.db.Query(ctx, `
		SELECT type, amount::text, balance_before::text, balance_after::text,
		       COALESCE(base_cost::text,'0'), COALESCE(rate_multiplier::text,'0'), remark, created_at
		FROM balance_ledger
		ORDER BY created_at DESC
		LIMIT 200
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanLedger(rows)
}

type rowsScanner interface {
	Scan(dest ...any) error
	Next() bool
	Err() error
}

func scanGatewayLogs(rows rowsScanner) ([]GatewayLog, error) {
	items := []GatewayLog{}
	for rows.Next() {
		var item GatewayLog
		if err := rows.Scan(&item.RequestID, &item.ModelID, &item.Status, &item.HTTPStatus, &item.TotalTokens, &item.BaseCost, &item.Charge, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func scanLedger(rows rowsScanner) ([]LedgerRecord, error) {
	items := []LedgerRecord{}
	for rows.Next() {
		var item LedgerRecord
		if err := rows.Scan(&item.Type, &item.Amount, &item.BalanceBefore, &item.BalanceAfter, &item.BaseCost, &item.RateMultiplier, &item.Remark, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func scanDaily(rows rowsScanner) ([]DailyStat, error) {
	items := []DailyStat{}
	for rows.Next() {
		var item DailyStat
		if err := rows.Scan(&item.Day, &item.Requests, &item.Successes, &item.Failures, &item.TotalTokens, &item.BaseCost, &item.Charge, &item.GrossProfit); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func scanModelStats(rows rowsScanner) ([]ModelStat, error) {
	items := []ModelStat{}
	for rows.Next() {
		var item ModelStat
		if err := rows.Scan(&item.ModelID, &item.Requests, &item.TotalTokens, &item.BaseCost, &item.Charge); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
