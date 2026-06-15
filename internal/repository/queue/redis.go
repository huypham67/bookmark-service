package queue

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type redisPublisher struct {
	client *redis.Client
}

// NewRedisPublisher creates a new Redis-based queue publisher with the provided Redis client.
func NewRedisPublisher(client *redis.Client) Publisher {
	return &redisPublisher{
		client: client,
	}
}

// Enqueue pushes the given payloads onto the tail of the specified queue. If no payloads are provided, this is a no-op.
func (p *redisPublisher) Enqueue(ctx context.Context, queue string, payloads ...[]byte) error {
	if len(payloads) == 0 {
		return nil
	}

	values := make([]interface{}, len(payloads))
	for i, payload := range payloads {
		values[i] = payload
	}

	return p.client.RPush(ctx, queue, values...).Err()
}
