package job

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"

	"lingshu/backend/internal/repository"
	"lingshu/backend/internal/service"
)

type channelRepository interface {
	ListUnhealthy(ctx context.Context) ([]repository.Channel, error)
	MarkHealthy(ctx context.Context, id string) error
}

type channelTester interface {
	Test(ctx context.Context, id, baseURL string) (map[string]any, error)
}

type ChannelHealer struct {
	channels  channelRepository
	tester    channelTester
	redis     *redis.Client
	interval  time.Duration
	threshold int
}

func NewChannelHealer(channels repository.ChannelRepository, tester service.ChannelService, redisClient *redis.Client, intervalSeconds int, successThreshold int) *ChannelHealer {
	if intervalSeconds <= 0 {
		intervalSeconds = 300
	}
	if successThreshold <= 0 {
		successThreshold = 3
	}
	return &ChannelHealer{
		channels:  channels,
		tester:    tester,
		redis:     redisClient,
		interval:  time.Duration(intervalSeconds) * time.Second,
		threshold: successThreshold,
	}
}

func (h *ChannelHealer) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(h.interval)
		defer ticker.Stop()
		h.RunOnce(ctx)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				h.RunOnce(ctx)
			}
		}
	}()
}

func (h *ChannelHealer) RunOnce(ctx context.Context) {
	channels, err := h.channels.ListUnhealthy(ctx)
	if err != nil {
		log.Printf("list unhealthy channels: %v", err)
		return
	}
	for _, channel := range channels {
		result, err := h.tester.Test(ctx, channel.ID, channel.BaseURL)
		if err != nil {
			log.Printf("heal test channel %s: %v", channel.ID, err)
			h.resetSuccess(ctx, channel.ID)
			continue
		}
		ok, _ := result["ok"].(bool)
		if !ok {
			h.resetSuccess(ctx, channel.ID)
			continue
		}
		count := h.recordSuccess(ctx, channel.ID)
		if count >= int64(h.threshold) {
			if err := h.channels.MarkHealthy(ctx, channel.ID); err != nil {
				log.Printf("mark channel %s healthy: %v", channel.ID, err)
				continue
			}
			h.resetSuccess(ctx, channel.ID)
		}
	}
}

func (h *ChannelHealer) recordSuccess(ctx context.Context, channelID string) int64 {
	if h.redis == nil {
		return int64(h.threshold)
	}
	key := "channel_heal:" + channelID
	count, err := h.redis.Incr(ctx, key).Result()
	if err != nil {
		log.Printf("increment heal counter %s: %v", channelID, err)
		return 0
	}
	_ = h.redis.Expire(ctx, key, 30*time.Minute).Err()
	return count
}

func (h *ChannelHealer) resetSuccess(ctx context.Context, channelID string) {
	if h.redis == nil {
		return
	}
	_ = h.redis.Del(ctx, "channel_heal:"+channelID).Err()
}
