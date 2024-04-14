package app

import (
	"context"
	"fmt"
	"log"
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

func (r *RedisClient) AddPending(id int) error {
	return r.client.SAdd(r.ctx, "pending", id).Err()
}

func (r *RedisClient) IsPending(id int) bool {
	return r.client.SIsMember(r.ctx, "pending", id).Val()
}

func (r *RedisClient) RemovePending(id int) error {
	return r.client.SRem(r.ctx, "pending", id).Err()
}

func (r *RedisClient) SetSeq(key string, value interface{}) error {
	err := r.client.Set(r.ctx, key, value, 60*time.Second).Err()
	if err != nil {
		log.Printf("failed to set value: %w", err)
	}
	return nil
}

func (r *RedisClient) GetSeq(key string) (int, error) {
	val, err := r.client.Get(r.ctx, key).Int()
	if err != nil {
		log.Printf("failed to get value: %w", err)
	}
	return val, nil
}

func (r *RedisClient) RemoveSeq(key string) error {
	err := r.client.Del(r.ctx, key).Err()
	if err != nil {
		log.Printf("failed to del value: %w", err)
	}
	return nil
}
