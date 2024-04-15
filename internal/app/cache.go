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
	return r.client.Set(r.ctx, key, value, 60*time.Second).Err()
}

func (r *RedisClient) GetSeq(key string) (int, error) {
	return r.client.Get(r.ctx, key).Int()
}

func (r *RedisClient) RemoveSeq(key string) error {
	return r.client.Del(r.ctx, key).Err()
}
