package app

import (
	"context"
	"fmt"
	"strconv"
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
	idStr := strconv.Itoa(id)
	// return r.client.SAdd(r.ctx, "pending", idStr).Err()
	return r.client.Set(r.ctx, idStr, "pending", 60*time.Second).Err()
}

func (r *RedisClient) IsPending(id int) bool {
	idStr := strconv.Itoa(id)
	// return r.client.SIsMember(r.ctx, "pending", idStr).Val()
	val := r.client.Get(r.ctx, idStr).Val()
	return val == "pending"
}

func (r *RedisClient) RemovePending(id int) error {
	idStr := strconv.Itoa(id)
	// return r.client.SRem(r.ctx, "pending", idStr).Err()
	return r.client.Del(r.ctx, idStr).Err()
}

func (r *RedisClient) SetSeq(key string, value interface{}) error {
	return r.client.Set(r.ctx, key, value, 60*time.Second).Err()
}

func (r *RedisClient) GetSeq(key string) (string, error) {
	return r.client.Get(r.ctx, key).Result()
}

func (r *RedisClient) RemoveSeq(key string) error {
	return r.client.Del(r.ctx, key).Err()
}
