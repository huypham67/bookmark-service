package cache

import (
	"testing"

	"github.com/huypham67/bookmark-common/pkg/redis"
)

func newTestRepository(t *testing.T) (Repository, *redis.Mock) {
	t.Helper()

	mockRedis := redis.NewMock(t)
	repo := NewRedis(mockRedis.Client)

	return repo, mockRedis
}
