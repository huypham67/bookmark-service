package repository

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Link defines the contract for link repository operations.
// mockery --name=Link --dir=internal/repository --output=internal/repository/mocks --filename=link_repository.go
type Link interface {
	SaveLink(ctx context.Context, code string, url string, exp int64) error
	CheckExists(ctx context.Context, code string) (bool, error)
	GetLink(ctx context.Context, code string) (string, error)
}

type linkRepository struct {
	client *redis.Client
}

// NewLinkRepository creates a new link repository with the given Redis client.
func NewLinkRepository(client *redis.Client) Link {
	return &linkRepository{
		client: client,
	}
}

// SaveLink saves a shortened URL code mapping to the provided URL in Redis with an expiration time.
func (r *linkRepository) SaveLink(ctx context.Context, code string, url string, exp int64) error {
	return r.client.Set(ctx, code, url, time.Duration(exp)*time.Second).Err()
}

// CheckExists checks whether a shortened URL code already exists in Redis.
func (r *linkRepository) CheckExists(ctx context.Context, code string) (bool, error) {
	result, err := r.client.Exists(ctx, code).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// GetLink retrieves the original URL for a given shortened code from Redis.
func (r *linkRepository) GetLink(ctx context.Context, code string) (string, error) {
	url, err := r.client.Get(ctx, code).Result()
	if err != nil {
		return "", err
	}
	return url, nil
}
