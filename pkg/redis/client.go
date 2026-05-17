package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// NewRedisClient initializes and returns a new Redis client based on environment variables with the specified prefix.
func NewRedisClient(envPrefix string) (*redis.Client, error) {
	config, err := LoadRedisConfig(envPrefix)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(&redis.Options{
		Addr:         config.Addr,
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
		_ = client.Close()
		return nil, err
	}

	return client, nil
}
