package ping

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type redisPinger struct {
	client *redis.Client
}

// NewRedisPinger creates a new Pinger instance with the given Redis client (Redis implementation).
func NewRedisPinger(client *redis.Client) Pinger {
	return &redisPinger{
		client: client,
	}
}

// Ping checks the connectivity to the Redis server by sending a PING command and returns any error encountered.
func (p *redisPinger) Ping(ctx context.Context) error {
	return p.client.Ping(ctx).Err()
}
