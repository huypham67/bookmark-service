package cache

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrCacheMiss = errors.New("cache miss")
)

type redisRepository struct {
	client *redis.Client
}

// NewRedis creates a new RedisRepository with the provided Redis client.
func NewRedis(client *redis.Client) Repository {
	return &redisRepository{
		client: client,
	}
}

// SetCache stores a value in Redis under the specified hash and field keys with a given time-to-live (TTL).
// It uses a pipeline to execute both the HSet and Expire commands atomically.
func (r *redisRepository) SetCache(ctx context.Context, hashKey, fieldKey string, value []byte, ttl time.Duration) error {
	pipeline := r.client.Pipeline()
	pipeline.HSet(ctx, hashKey, fieldKey, value)
	pipeline.Expire(ctx, hashKey, ttl)
	_, err := pipeline.Exec(ctx)
	return err
}

// GetCache retrieves a value from Redis using the specified hash and field keys. If the key is not found, it returns an ErrCacheMiss error.
func (r *redisRepository) GetCache(ctx context.Context, hashKey, fieldKey string) ([]byte, error) {
	result, err := r.client.HGet(ctx, hashKey, fieldKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrCacheMiss
		}
		return nil, err
	}
	return result, nil
}

// DeleteCacheByFieldKey removes a specific field from a Redis hash.
func (r *redisRepository) DeleteCacheByFieldKey(ctx context.Context, hashKey, fieldKey string) error {
	return r.client.HDel(ctx, hashKey, fieldKey).Err()
}

// DeleteCacheByHashKey deletes an entire Redis hash, effectively removing all cached entries under that hash key.
func (r *redisRepository) DeleteCacheByHashKey(ctx context.Context, hashKey string) error {
	return r.client.Del(ctx, hashKey).Err()
}
