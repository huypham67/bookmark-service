package cache

import (
	"context"
	"time"
)

// Repository defines the interface for cache operations, such as setting and getting cached values.
//
//go:generate mockery --name=Repository --output=./mocks --outpkg=mocks --filename=mock_repo.go
type Repository interface {
	SetCache(ctx context.Context, hashKey, fieldKey string, value []byte, ttl time.Duration) error
	GetCache(ctx context.Context, hashKey, fieldKey string) ([]byte, error)
	DeleteCacheByFieldKey(ctx context.Context, hashKey, fieldKey string) error
	DeleteCacheByHashKey(ctx context.Context, hashKey string) error
}
