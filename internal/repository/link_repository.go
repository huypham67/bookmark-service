package repository

import (
	"time"

	"github.com/huypham67/bookmark-management/infrastructure/redis"
)

// Link defines the contract for link repository operations.
type Link interface {
	SaveLink(code string, url string, exp int64) error
	CheckExists(code string) (bool, error)
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
	return r.redisClient.Set(code, url, time.Duration(exp)*time.Second)
}

// CheckExists checks whether a shortened URL code already exists in Redis.
func (r *linkRepository) CheckExists(code string) (bool, error) {
	return r.redisClient.Exists(code)
}
