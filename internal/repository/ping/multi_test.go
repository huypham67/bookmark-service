package ping

import (
	"context"
	"errors"
	"testing"

	"github.com/huypham67/bookmark-service/internal/repository/ping/mocks"
	"github.com/stretchr/testify/require"
)

func TestMultiPinger_Ping(t *testing.T) {
	t.Parallel()

	errUnhealthy := errors.New("dependency unhealthy")

	testCases := []struct {
		name        string
		setup       func(*testing.T, context.Context) []Pinger
		wantErr     error
		verifyMocks func(*testing.T, context.Context, []Pinger)
	}{
		{
			name: "should return nil when all dependencies are healthy",
			setup: func(t *testing.T, ctx context.Context) []Pinger {
				first := newMockPinger(t)
				second := newMockPinger(t)
				first.On("Ping", ctx).Return(nil).Once()
				second.On("Ping", ctx).Return(nil).Once()

				return []Pinger{first, second}
			},
			wantErr: nil,
		},
		{
			name: "should fail fast and skip later dependencies when the first is unhealthy",
			setup: func(t *testing.T, ctx context.Context) []Pinger {
				first := newMockPinger(t)
				second := newMockPinger(t)
				first.On("Ping", ctx).Return(errUnhealthy).Once()

				return []Pinger{first, second}
			},
			wantErr: errUnhealthy,
			verifyMocks: func(t *testing.T, ctx context.Context, pingers []Pinger) {
				second := pingers[1].(*mocks.Pinger)
				second.AssertNotCalled(t, "Ping", ctx)
			},
		},
		{
			name: "should return the error when a later dependency is unhealthy",
			setup: func(t *testing.T, ctx context.Context) []Pinger {
				first := newMockPinger(t)
				second := newMockPinger(t)
				first.On("Ping", ctx).Return(nil).Once()
				second.On("Ping", ctx).Return(errUnhealthy).Once()

				return []Pinger{first, second}
			},
			wantErr: errUnhealthy,
		},
		{
			name: "should return nil when there are no dependencies",
			setup: func(t *testing.T, ctx context.Context) []Pinger {
				return nil
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			pingers := tc.setup(t, ctx)
			pinger := NewMulti(pingers...)

			err := pinger.Ping(ctx)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
			}

			if tc.verifyMocks != nil {
				tc.verifyMocks(t, ctx, pingers)
			}
		})
	}
}
