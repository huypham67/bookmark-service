package ping

import (
	"testing"

	"github.com/huypham67/bookmark-common/pkg/redis"
)

func newTestPinger(t *testing.T) (Pinger, *redis.Mock) {
	t.Helper()

	mockRedis := redis.NewMock(t)
	pinger := NewRedis(mockRedis.Client)

	return pinger, mockRedis
}
