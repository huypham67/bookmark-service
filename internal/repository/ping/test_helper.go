package ping

import (
	"testing"

	"github.com/huypham67/bookmark-common/pkg/redis"
	"github.com/huypham67/bookmark-common/pkg/sqldb"
	"github.com/huypham67/bookmark-service/internal/repository/ping/mocks"
	"gorm.io/gorm"
)

func newTestRedisPinger(t *testing.T) (Pinger, *redis.Mock) {
	t.Helper()

	mockRedis := redis.NewMock(t)
	pinger := NewRedis(mockRedis.Client)

	return pinger, mockRedis
}

func newTestSQLDBPinger(t *testing.T) (Pinger, *gorm.DB) {
	t.Helper()

	db := sqldb.NewMock(t)
	pinger := NewSQLDB(db)

	return pinger, db
}

func newMockPinger(t *testing.T) *mocks.Pinger {
	t.Helper()

	m := &mocks.Pinger{}
	t.Cleanup(func() {
		m.AssertExpectations(t)
	})

	return m
}
