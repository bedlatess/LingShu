package job

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"lingshu/backend/internal/repository"
)

type fakeChannelHealerRepo struct {
	channels  []repository.Channel
	healthyID string
}

func (r *fakeChannelHealerRepo) ListUnhealthy(ctx context.Context) ([]repository.Channel, error) {
	return r.channels, nil
}

func (r *fakeChannelHealerRepo) MarkHealthy(ctx context.Context, id string) error {
	r.healthyID = id
	return nil
}

type fakeChannelTester struct {
	ok bool
}

func (t fakeChannelTester) Test(ctx context.Context, id, baseURL string) (map[string]any, error) {
	return map[string]any{"ok": t.ok}, nil
}

func TestChannelHealerMarksHealthyAfterThreeSuccesses(t *testing.T) {
	ctx := context.Background()
	redisServer := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	repo := &fakeChannelHealerRepo{channels: []repository.Channel{{
		ID:      "channel-1",
		BaseURL: "http://upstream.local",
		Health:  "unhealthy",
	}}}
	healer := &ChannelHealer{
		channels:  repo,
		tester:    fakeChannelTester{ok: true},
		redis:     client,
		threshold: 3,
	}

	healer.RunOnce(ctx)
	if repo.healthyID != "" {
		t.Fatalf("channel healed after first success")
	}
	healer.RunOnce(ctx)
	if repo.healthyID != "" {
		t.Fatalf("channel healed after second success")
	}
	healer.RunOnce(ctx)
	if repo.healthyID != "channel-1" {
		t.Fatalf("healthyID=%q want channel-1", repo.healthyID)
	}
	if value, err := redisServer.Get("channel_heal:channel-1"); err == nil || value != "" {
		t.Fatalf("heal counter not reset: value=%q err=%v", value, err)
	}
}
