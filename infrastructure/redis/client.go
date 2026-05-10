package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient wraps the go-redis client with methods for Redis operations.
type RedisClient struct {
	client *redis.Client
}

// NewRedisClient initializes a new Redis client with the provided configuration and tests the connection.
func NewRedisClient(config RedisConfig) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", config.Host, config.Port),
		Password:     config.Password,
		DB:           config.Database,
		PoolSize:     10,
		MinIdleConns: 5,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisClient{client: client}, nil
}

// Set stores a value at the given key with optional expiration time.
func (rc *RedisClient) Set(key string, value interface{}, expiration time.Duration) error {
	return rc.client.Set(context.Background(), key, value, expiration).Err()
}

// Get retrieves the value associated with the given key from Redis.
func (rc *RedisClient) Get(key string) (string, error) {
	return rc.client.Get(context.Background(), key).Result()
}

// Exists checks if a key exists in Redis. Returns true if the key exists, false otherwise.
func (rc *RedisClient) Exists(key string) (bool, error) {
	result, err := rc.client.Exists(context.Background(), key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// Del removes a key from Redis.
func (rc *RedisClient) Del(key string) error {
	return rc.client.Del(context.Background(), key).Err()
}

// Ping checks the connection to Redis by sending a PING command.
func (rc *RedisClient) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return rc.client.Ping(ctx).Err()
}
