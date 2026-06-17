package billing

import (
	"context"
	"sync"
	"testing"
)

type memoryFrozenStore struct {
	mu   sync.Mutex
	data map[string]int64
}

func (s *memoryFrozenStore) Add(_ context.Context, userID string, delta int64) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.data == nil {
		s.data = map[string]int64{}
	}
	s.data[userID] += delta
	return s.data[userID], nil
}

func (s *memoryFrozenStore) Get(_ context.Context, userID string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.data[userID], nil
}

func TestReserveConcurrentDoesNotOverFreeze(t *testing.T) {
	store := &memoryFrozenStore{}
	ctx := context.Background()
	const balance = int64(10_000_000)
	const estimate = int64(1_000_000)

	var wg sync.WaitGroup
	successes := make(chan bool, 50)
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := Reserve(ctx, store, "u1", balance, estimate)
			successes <- err == nil
		}()
	}
	wg.Wait()
	close(successes)

	count := 0
	for ok := range successes {
		if ok {
			count++
		}
	}
	if count != 10 {
		t.Fatalf("successes = %d, want 10", count)
	}
	frozen, _ := store.Get(ctx, "u1")
	if frozen != balance {
		t.Fatalf("frozen = %d, want %d", frozen, balance)
	}
}

func TestDecimalStringToUnitsAvoidsFloat(t *testing.T) {
	got, err := DecimalStringToUnits("1.234567")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got != 1_234_567 {
		t.Fatalf("units = %d, want 1234567", got)
	}
	rendered := UnitsToDecimalString(got)
	if rendered != "1.234567" {
		t.Fatalf("rendered = %q", rendered)
	}
}
