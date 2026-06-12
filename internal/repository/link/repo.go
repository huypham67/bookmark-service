package link

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// Repository defines the contract for link repositories.
//
//go:generate mockery --name=Repository --output=./mocks --outpkg=mocks --filename=mock_repo.go
type Repository interface {
	SaveLink(ctx context.Context, code string, url string, exp int64) error
	GetLink(ctx context.Context, code string) (string, error)
	CheckExists(ctx context.Context, code string) (bool, error)
}

type repository struct {
	client *redis.Client
}

// NewRepository creates a new instance of the link repository with the provided Redis client.
func NewRepository(client *redis.Client) Repository {
	return &repository{
		client: client,
	}
}
