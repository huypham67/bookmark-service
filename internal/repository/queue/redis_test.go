package queue

import (
	"context"
	"testing"

	pkgRedis "github.com/huypham67/bookmark-common/pkg/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPublisher_Enqueue(t *testing.T) {
	t.Parallel()

	const queueName = "bookmark:import:jobs"

	// Payloads are opaque bytes at this layer: the publisher only RPUSHes them.
	testCases := []struct {
		name     string
		payloads [][]byte
		setup    func(*pkgRedis.Mock)
		verify   func(*testing.T, *pkgRedis.Mock, error)
	}{
		{
			name:     "pushes a single payload to the tail",
			payloads: [][]byte{[]byte("a")},
			verify: func(t *testing.T, mockRedis *pkgRedis.Mock, err error) {
				require.NoError(t, err)
				got, lerr := mockRedis.Server.List(queueName)
				require.NoError(t, lerr)
				assert.Equal(t, []string{"a"}, got)
			},
		},
		{
			name:     "pushes multiple payloads in order in one call",
			payloads: [][]byte{[]byte("a"), []byte("b"), []byte("c")},
			verify: func(t *testing.T, mockRedis *pkgRedis.Mock, err error) {
				require.NoError(t, err)
				got, lerr := mockRedis.Server.List(queueName)
				require.NoError(t, lerr)
				assert.Equal(t, []string{"a", "b", "c"}, got)
			},
		},
		{
			name:     "no payloads is a no-op",
			payloads: nil,
			verify: func(t *testing.T, mockRedis *pkgRedis.Mock, err error) {
				require.NoError(t, err)
				assert.False(t, mockRedis.Server.Exists(queueName), "queue should not be created")
			},
		},
		{
			name:     "returns error when Redis client is closed",
			payloads: [][]byte{[]byte("a")},
			setup: func(mockRedis *pkgRedis.Mock) {
				mockRedis.Close()
			},
			verify: func(t *testing.T, _ *pkgRedis.Mock, err error) {
				assert.Error(t, err)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			pub, mockRedis := newTestRedisPublisher(t)
			if tc.setup != nil {
				tc.setup(mockRedis)
			}

			err := pub.Enqueue(ctx, queueName, tc.payloads...)

			tc.verify(t, mockRedis, err)
		})
	}
}
