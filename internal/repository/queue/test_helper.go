package queue

import (
	"testing"

	"github.com/huypham67/bookmark-common/pkg/redis"
)

func newTestRedisPublisher(t *testing.T) (Publisher, *redis.Mock) {
	t.Helper()

	mockRedis := redis.NewMock(t)
	pub := NewRedisPublisher(mockRedis.Client)

	return pub, mockRedis
}
