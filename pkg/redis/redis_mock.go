package redis

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// MockRedis provides a mock Redis server and client for testing purposes.
type MockRedis struct {
	Server *miniredis.Miniredis
	Client *redis.Client
}

// NewMockRedis creates a new mock Redis server and client for testing purposes.
func NewMockRedis(t *testing.T) *MockRedis {
	t.Helper()

	server, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start mock Redis server: %v", err)
	}

	t.Cleanup(func() {
		server.Close()
	})

	client := redis.NewClient(&redis.Options{
		Addr: server.Addr(),
	})

	return &MockRedis{
		Server: server,
		Client: client,
	}
}

// FastForward advances the mock Redis server's internal clock by the specified duration, allowing tests to simulate time-based behavior such as key expiration.
func (m *MockRedis) FastForward(duration time.Duration) {
	m.Server.FastForward(duration)
}
