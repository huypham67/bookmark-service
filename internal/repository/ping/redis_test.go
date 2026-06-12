package ping

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisPinger_Ping(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		verify func(*testing.T, error)
	}{
		{
			name: "should successfully ping Redis server",
			verify: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "should return error when Redis client is closed",
			verify: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			pinger, mockRedis := newTestRedisPinger(t)

			if tc.name == "should return error when Redis client is closed" {
				mockRedis.Close()
			}

			err := pinger.Ping(ctx)

			tc.verify(t, err)
		})
	}
}
