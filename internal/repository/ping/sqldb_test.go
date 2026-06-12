package ping

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestSQLDBPinger_Ping(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		setup  func(*testing.T) Pinger
		verify func(*testing.T, error)
	}{
		{
			name: "should successfully ping the database",
			setup: func(t *testing.T) Pinger {
				pinger, _ := newTestSQLDBPinger(t)
				return pinger
			},
			verify: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "should return error when the database connection is closed",
			setup: func(t *testing.T) Pinger {
				pinger, db := newTestSQLDBPinger(t)

				sqlDB, err := db.DB()
				require.NoError(t, err)
				require.NoError(t, sqlDB.Close())

				return pinger
			},
			verify: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "should return error when the gorm DB has no valid connection pool",
			setup: func(t *testing.T) Pinger {
				return NewSQLDB(&gorm.DB{Config: &gorm.Config{}})
			},
			verify: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			pinger := tc.setup(t)

			err := pinger.Ping(ctx)

			tc.verify(t, err)
		})
	}
}
