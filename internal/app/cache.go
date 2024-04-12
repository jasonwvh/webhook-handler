package app

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisClient struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisClient(host string) *RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:6379", host),
	})
	return &RedisClient{
		client: rdb,
		ctx:    context.Background(),
	}
}

func (r *RedisClient) SetValue(key string, value interface{}) error {
	err := r.client.Set(r.ctx, key, value, 60*time.Second).Err()
	if err != nil {
		return fmt.Errorf("failed to set value: %w", err)
	}
	return nil
}

func (r *RedisClient) GetValue(key string) (string, error) {
	val, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("failed to get value: %w", err)
	}
	return val, nil
}

func (r *RedisClient) RemoveKey(key string) error {
	err := r.client.Del(r.ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to remove key: %w", err)
	}
	return nil
}
