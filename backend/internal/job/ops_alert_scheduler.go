package job

import (
	"context"
	"log"
	"time"
)

type opsAlertEvaluator interface {
	Evaluate(ctx context.Context) error
}

type OpsAlertScheduler struct {
	alerts   opsAlertEvaluator
	interval time.Duration
}

func NewOpsAlertScheduler(alerts opsAlertEvaluator, interval time.Duration) *OpsAlertScheduler {
	if interval <= 0 {
		interval = time.Minute
	}
	return &OpsAlertScheduler{alerts: alerts, interval: interval}
}

func (s *OpsAlertScheduler) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()
		s.RunOnce(ctx)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.RunOnce(ctx)
			}
		}
	}()
}

func (s *OpsAlertScheduler) RunOnce(ctx context.Context) {
	if err := s.alerts.Evaluate(ctx); err != nil {
		log.Printf("evaluate ops alerts: %v", err)
	}
}
