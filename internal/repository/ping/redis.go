package ping

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type redisPinger struct {
	client *redis.Client
}

// NewRedis creates a new RedisPinger with the provided Redis client.
func NewRedis(client *redis.Client) Pinger {
	return &redisPinger{
		client: client,
	}
}

// Ping checks the health of the Redis connection by sending a PING command and returning any error encountered.
func (p *redisPinger) Ping(ctx context.Context) error {
	return p.client.Ping(ctx).Err()
}
