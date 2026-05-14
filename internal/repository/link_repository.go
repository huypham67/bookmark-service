package repository

import (
	"context"
	"time"

	"github.com/huypham67/bookmark-management/pkg/redis"
)

// Link defines the contract for link repository operations.
type Link interface {
	SaveLink(code string, url string, exp int64) error
	CheckExists(code string) (bool, error)
	GetLink(code string) (string, error)
}

type linkRepository struct {
	redisClient *redis.RedisClient
}

// NewLinkRepository creates a new link repository with the given Redis client.
func NewLinkRepository(redisClient *redis.RedisClient) Link {
	return &linkRepository{
		redisClient: redisClient,
	}
}

// SaveLink saves a shortened URL code mapping to the provided URL in Redis with an expiration time.
func (r *linkRepository) SaveLink(code string, url string, exp int64) error {
	return r.redisClient.Client.Set(context.Background(), code, url, time.Duration(exp)*time.Second).Err()
}

// CheckExists checks whether a shortened URL code already exists in Redis.
func (r *linkRepository) CheckExists(code string) (bool, error) {
	result, err := r.redisClient.Client.Exists(context.Background(), code).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// GetLink retrieves the original URL for a given shortened code from Redis.
func (r *linkRepository) GetLink(code string) (string, error) {
	url, err := r.redisClient.Client.Get(context.Background(), code).Result()
	if err != nil {
		return "", err
	}
	return url, nil
}
