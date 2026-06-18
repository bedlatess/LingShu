package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type GatewayLog struct {
	RequestID   string    `json:"request_id"`
	UserID      string    `json:"user_id,omitempty"`
	ModelID     string    `json:"model_id"`
	Status      string    `json:"status"`
	HTTPStatus  int       `json:"http_status"`
	TotalTokens int       `json:"total_tokens"`
	BaseCost    string    `json:"base_cost"`
	Charge      string    `json:"charge"`
	CreatedAt   time.Time `json:"created_at"`
}

type LedgerRecord struct {
	UserID         string    `json:"user_id,omitempty"`
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
	TotalCharge   string `json:"total_charge"`
	TotalRecharge string `json:"total_recharge"`
}

type UserSummaryStats struct {
	TotalCharge   string `json:"total_charge"`
	TotalRecharge string `json:"total_recharge"`
}

type ReportRepository struct {
	db *pgxpool.Pool
}

type UserLogFilter struct {
	Status string
	Model  string
	From   string
	To     string
}

type UserLedgerFilter struct {
	Type string
	From string
	To   string
}

type ReportRow struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Requests    int    `json:"requests"`
	Successes   int    `json:"successes"`
	Failures    int    `json:"failures"`
	BaseCost    string `json:"base_cost"`
	Charge      string `json:"charge"`
	GrossProfit string `json:"gross_profit"`
}

func NewReportRepository(db *pgxpool.Pool) ReportRepository {
	return ReportRepository{db: db}
}

func (r ReportRepository) UserLogs(ctx context.Context, userID string) ([]GatewayLog, error) {
	items, _, err := r.UserLogsPaged(ctx, userID, 100, 0)
	return items, err
}

func (r ReportRepository) UserLogsPaged(ctx context.Context, userID string, limit, offset int) ([]GatewayLog, int, error) {
	return r.UserLogsFilteredPaged(ctx, userID, UserLogFilter{}, limit, offset)
}

func (r ReportRepository) UserLogsFilteredPaged(ctx context.Context, userID string, filter UserLogFilter, limit, offset int) ([]GatewayLog, int, error) {
	where, args := userLogWhere(userID, filter)
	var total int
	if err := r.db.QueryRow(ctx, `SELECT count(*)::int FROM gateway_requests `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, limit, offset)
	rows, err := r.db.Query(ctx, `
		SELECT request_id, user_id::text, COALESCE(model_id::text,''), status, http_status, total_tokens, base_cost::text, charge::text, created_at
		FROM gateway_requests
		`+where+`
		ORDER BY created_at DESC
		LIMIT $`+fmt.Sprint(len(args)-1)+` OFFSET $`+fmt.Sprint(len(args))+`
	`, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items, err := scanGatewayLogs(rows)
	return items, total, err
}

func (r ReportRepository) UserLedger(ctx context.Context, userID string) ([]LedgerRecord, error) {
	items, _, err := r.UserLedgerPaged(ctx, userID, 100, 0)
	return items, err
}

func (r ReportRepository) UserLedgerPaged(ctx context.Context, userID string, limit, offset int) ([]LedgerRecord, int, error) {
	return r.UserLedgerFilteredPaged(ctx, userID, UserLedgerFilter{}, limit, offset)
}

func (r ReportRepository) UserLedgerFilteredPaged(ctx context.Context, userID string, filter UserLedgerFilter, limit, offset int) ([]LedgerRecord, int, error) {
	where, args := userLedgerWhere(userID, filter)
	var total int
	if err := r.db.QueryRow(ctx, `SELECT count(*)::int FROM balance_ledger `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, limit, offset)
	rows, err := r.db.Query(ctx, `
		SELECT user_id::text, type, amount::text, balance_before::text, balance_after::text,
		       COALESCE(base_cost::text,'0'), COALESCE(rate_multiplier::text,'0'), remark, created_at
		FROM balance_ledger
		`+where+`
		ORDER BY created_at DESC
		LIMIT $`+fmt.Sprint(len(args)-1)+` OFFSET $`+fmt.Sprint(len(args))+`
	`, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items, err := scanLedger(rows)
	return items, total, err
}

func userLogWhere(userID string, filter UserLogFilter) (string, []any) {
	args := []any{userID}
	where := "WHERE user_id=$1"
	if filter.Status != "" && filter.Status != "all" {
		args = append(args, filter.Status)
		where += fmt.Sprintf(" AND status=$%d", len(args))
	}
	if filter.Model != "" {
		args = append(args, "%"+filter.Model+"%")
		where += fmt.Sprintf(" AND COALESCE(model_id::text,'') ILIKE $%d", len(args))
	}
	if filter.From != "" {
		args = append(args, filter.From)
		where += fmt.Sprintf(" AND created_at >= $%d::date", len(args))
	}
	if filter.To != "" {
		args = append(args, filter.To)
		where += fmt.Sprintf(" AND created_at < ($%d::date + interval '1 day')", len(args))
	}
	return where, args
}

func userLedgerWhere(userID string, filter UserLedgerFilter) (string, []any) {
	args := []any{userID}
	where := "WHERE user_id=$1"
	if filter.Type != "" && filter.Type != "all" {
		args = append(args, filter.Type)
		where += fmt.Sprintf(" AND type=$%d", len(args))
	}
	if filter.From != "" {
		args = append(args, filter.From)
		where += fmt.Sprintf(" AND created_at >= $%d::date", len(args))
	}
	if filter.To != "" {
		args = append(args, filter.To)
		where += fmt.Sprintf(" AND created_at < ($%d::date + interval '1 day')", len(args))
	}
	return where, args
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
		  COALESCE((SELECT sum(charge) FROM gateway_requests WHERE user_id=$1 AND date_trunc('month', created_at)=date_trunc('month', now())),0)::text,
		  COALESCE((SELECT sum(charge) FROM gateway_requests WHERE user_id=$1),0)::text,
		  COALESCE((SELECT sum(amount) FROM balance_ledger WHERE user_id=$1 AND amount > 0),0)::text
	`, userID).Scan(&item.TodayRequests, &item.TodayCharge, &item.MonthCharge, &item.TotalCharge, &item.TotalRecharge)
	return item, err
}

func (r ReportRepository) UserSummary(ctx context.Context, userID string) (UserSummaryStats, error) {
	var item UserSummaryStats
	err := r.db.QueryRow(ctx, `
		SELECT
		  COALESCE((SELECT sum(charge) FROM gateway_requests WHERE user_id=$1),0)::text,
		  COALESCE((SELECT sum(amount) FROM balance_ledger WHERE user_id=$1 AND amount > 0),0)::text
	`, userID).Scan(&item.TotalCharge, &item.TotalRecharge)
	return item, err
}

func (r ReportRepository) AdminLogs(ctx context.Context) ([]GatewayLog, error) {
	items, _, err := r.AdminLogsPaged(ctx, 100, 0)
	return items, err
}

func (r ReportRepository) AdminLogsPaged(ctx context.Context, limit, offset int) ([]GatewayLog, int, error) {
	var total int
	if err := r.db.QueryRow(ctx, `SELECT count(*)::int FROM gateway_requests`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.Query(ctx, `
		SELECT request_id, user_id::text, COALESCE(model_id::text,''), status, http_status, total_tokens, base_cost::text, charge::text, created_at
		FROM gateway_requests
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items, err := scanGatewayLogs(rows)
	return items, total, err
}

func (r ReportRepository) AdminLedger(ctx context.Context) ([]LedgerRecord, error) {
	items, _, err := r.AdminLedgerPaged(ctx, 100, 0)
	return items, err
}

func (r ReportRepository) AdminLedgerPaged(ctx context.Context, limit, offset int) ([]LedgerRecord, int, error) {
	var total int
	if err := r.db.QueryRow(ctx, `SELECT count(*)::int FROM balance_ledger`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.Query(ctx, `
		SELECT user_id::text, type, amount::text, balance_before::text, balance_after::text,
		       COALESCE(base_cost::text,'0'), COALESCE(rate_multiplier::text,'0'), remark, created_at
		FROM balance_ledger
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items, err := scanLedger(rows)
	return items, total, err
}

func (r ReportRepository) ReportDaily(ctx context.Context, from, to string) ([]ReportRow, error) {
	return r.reportGroup(ctx, from, to, `
		SELECT to_char(date_trunc('day', created_at), 'YYYY-MM-DD') AS key,
		       to_char(date_trunc('day', created_at), 'YYYY-MM-DD') AS label,
		       count(*)::int,
		       count(*) FILTER (WHERE status='success')::int,
		       count(*) FILTER (WHERE status!='success')::int,
		       COALESCE(sum(base_cost),0)::text,
		       COALESCE(sum(charge),0)::text,
		       COALESCE(sum(charge - base_cost),0)::text
		FROM gateway_requests
	`, "created_at", "GROUP BY 1,2 ORDER BY 1 DESC")
}

func (r ReportRepository) ReportByUser(ctx context.Context, from, to string) ([]ReportRow, error) {
	return r.reportGroup(ctx, from, to, `
		SELECT u.id::text AS key,
		       u.username AS label,
		       count(*)::int,
		       count(*) FILTER (WHERE gr.status='success')::int,
		       count(*) FILTER (WHERE gr.status!='success')::int,
		       COALESCE(sum(gr.base_cost),0)::text,
		       COALESCE(sum(gr.charge),0)::text,
		       COALESCE(sum(gr.charge - gr.base_cost),0)::text
		FROM gateway_requests gr
		JOIN users u ON u.id = gr.user_id
	`, "gr.created_at", "GROUP BY 1,2 ORDER BY sum(gr.charge) DESC")
}

func (r ReportRepository) ReportByModel(ctx context.Context, from, to string) ([]ReportRow, error) {
	return r.reportGroup(ctx, from, to, `
		SELECT COALESCE(m.id::text, '') AS key,
		       COALESCE(m.public_name, '未绑定模型') AS label,
		       count(*)::int,
		       count(*) FILTER (WHERE gr.status='success')::int,
		       count(*) FILTER (WHERE gr.status!='success')::int,
		       COALESCE(sum(gr.base_cost),0)::text,
		       COALESCE(sum(gr.charge),0)::text,
		       COALESCE(sum(gr.charge - gr.base_cost),0)::text
		FROM gateway_requests gr
		LEFT JOIN models m ON m.id = gr.model_id
	`, "gr.created_at", "GROUP BY 1,2 ORDER BY sum(gr.charge) DESC")
}

func (r ReportRepository) ReportByChannel(ctx context.Context, from, to string) ([]ReportRow, error) {
	return r.reportGroup(ctx, from, to, `
		SELECT COALESCE(c.id::text, '') AS key,
		       COALESCE(c.name, '未绑定渠道') AS label,
		       count(*)::int,
		       count(*) FILTER (WHERE gr.status='success')::int,
		       count(*) FILTER (WHERE gr.status!='success')::int,
		       COALESCE(sum(gr.base_cost),0)::text,
		       COALESCE(sum(gr.charge),0)::text,
		       COALESCE(sum(gr.charge - gr.base_cost),0)::text
		FROM gateway_requests gr
		LEFT JOIN upstream_channels c ON c.id = gr.channel_id
	`, "gr.created_at", "GROUP BY 1,2 ORDER BY sum(gr.charge) DESC")
}

func (r ReportRepository) reportGroup(ctx context.Context, from, to, selectClause, timeColumn, tail string) ([]ReportRow, error) {
	args := []any{}
	where := " WHERE 1=1"
	if from != "" {
		args = append(args, from)
		where += fmt.Sprintf(" AND %s >= $%d::date", timeColumn, len(args))
	}
	if to != "" {
		args = append(args, to)
		where += fmt.Sprintf(" AND %s < ($%d::date + interval '1 day')", timeColumn, len(args))
	}
	rows, err := r.db.Query(ctx, selectClause+where+" "+tail, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ReportRow{}
	for rows.Next() {
		var item ReportRow
		if err := rows.Scan(&item.Key, &item.Label, &item.Requests, &item.Successes, &item.Failures, &item.BaseCost, &item.Charge, &item.GrossProfit); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
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
		if err := rows.Scan(&item.RequestID, &item.UserID, &item.ModelID, &item.Status, &item.HTTPStatus, &item.TotalTokens, &item.BaseCost, &item.Charge, &item.CreatedAt); err != nil {
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
		if err := rows.Scan(&item.UserID, &item.Type, &item.Amount, &item.BalanceBefore, &item.BalanceAfter, &item.BaseCost, &item.RateMultiplier, &item.Remark, &item.CreatedAt); err != nil {
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
