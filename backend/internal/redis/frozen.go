package redisstore

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type FrozenStore struct {
	client *redis.Client
}

func NewFrozenStore(client *redis.Client) FrozenStore {
	return FrozenStore{client: client}
}

func (s FrozenStore) Add(ctx context.Context, userID string, delta int64) (int64, error) {
	return s.client.IncrBy(ctx, "frozen:"+userID, delta).Result()
}

func (s FrozenStore) Get(ctx context.Context, userID string) (int64, error) {
	value, err := s.client.Get(ctx, "frozen:"+userID).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return value, err
}

func (s FrozenStore) AllowRPM(ctx context.Context, subject string, limit int) (bool, error) {
	if limit <= 0 {
		return true, nil
	}
	key := "rpm:" + subject + ":" + time.Now().UTC().Format("200601021504")
	count, err := s.client.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}
	if count == 1 {
		_ = s.client.Expire(ctx, key, 2*time.Minute).Err()
	}
	return count <= int64(limit), nil
}

func (s FrozenStore) TryAcquire(ctx context.Context, subject string, limit int, ttl time.Duration) (bool, error) {
	if limit <= 0 {
		return true, nil
	}
	key := "concurrency:" + subject
	count, err := s.client.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}
	if count == 1 && ttl > 0 {
		_ = s.client.Expire(ctx, key, ttl).Err()
	}
	if count > int64(limit) {
		_, _ = s.client.Decr(ctx, key).Result()
		return false, nil
	}
	return true, nil
}

func (s FrozenStore) ReleaseConcurrency(ctx context.Context, subject string) {
	_, _ = s.client.Decr(ctx, "concurrency:"+subject).Result()
}

func (s FrozenStore) GetStickyChannel(ctx context.Context, key string) (string, error) {
	value, err := s.client.Get(ctx, "sticky:"+key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return value, err
}

func (s FrozenStore) SetStickyChannel(ctx context.Context, key, channelID string, ttl time.Duration) error {
	if key == "" || channelID == "" {
		return nil
	}
	return s.client.Set(ctx, "sticky:"+key, channelID, ttl).Err()
}
