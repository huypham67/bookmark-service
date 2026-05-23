package ping

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// Pinger defines the contract for a service that can check the health of a data store (e.g., Redis).
type Pinger interface {
	Ping(ctx context.Context) error
}

// NewPinger creates a new Pinger instance using the provided Redis client. It returns an implementation of the Pinger interface that can be used to check the health of the Redis connection.
func NewPinger(client *redis.Client) Pinger {
	return NewRedisPinger(client)
}
