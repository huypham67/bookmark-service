package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Pinger is an interface that defines methods to check Redis connectivity.
type Pinger interface {
	Ping() error
}

type pinger struct {
	client *redis.Client
}

// NewPinger creates a new Pinger instance with the given Redis client.
func NewPinger(client *redis.Client) Pinger {
	return &pinger{
		client: client,
	}
}

// Ping checks the connectivity to the Redis server by sending a PING command and returns any error encountered.
func (p *pinger) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return p.client.Ping(ctx).Err()
}
