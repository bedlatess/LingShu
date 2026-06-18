package job

import (
	"context"
	"log"
	"sync"
	"time"
)

type Scheduler struct {
	cleaner Cleaner
	mu      sync.Mutex
	lastRun string
}

func NewScheduler(cleaner Cleaner) *Scheduler {
	return &Scheduler{cleaner: cleaner}
}

func (s *Scheduler) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		s.maybeRun(ctx, time.Now())
		for {
			select {
			case <-ctx.Done():
				return
			case now := <-ticker.C:
				s.maybeRun(ctx, now)
			}
		}
	}()
}

func (s *Scheduler) maybeRun(ctx context.Context, now time.Time) {
	if now.Hour() != 3 {
		return
	}
	today := now.Format("2006-01-02")
	s.mu.Lock()
	if s.lastRun == today {
		s.mu.Unlock()
		return
	}
	s.lastRun = today
	s.mu.Unlock()

	results := s.cleaner.Run(ctx)
	if err := s.cleaner.SaveHistory(ctx, results); err != nil {
		log.Printf("save cleanup history: %v", err)
	}
}
