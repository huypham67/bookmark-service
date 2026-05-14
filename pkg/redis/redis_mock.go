package redis

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

type MockRedis struct {
	Server *miniredis.Miniredis
	Client *redis.Client
}

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
