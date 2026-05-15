package repository

import (
	"context"
	"testing"
	"time"

	"github.com/huypham67/bookmark-management/pkg/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRepository(
	t *testing.T,
) (Link, *redis.MockRedis) {
	t.Helper()

	mockRedis := redis.NewMockRedis(t)

	client := &redis.RedisClient{
		Client: mockRedis.Client,
	}

	repo := NewLinkRepository(client)

	return repo, mockRedis
}

func TestLinkRepository_SaveLink(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	type args struct {
		code string
		url  string
		exp  int64
	}

	testCases := []struct {
		name   string
		args   args
		verify func(*testing.T, Link, *redis.MockRedis, args)
	}{
		{
			name: "should save link successfully",
			args: args{
				code: "abc1234",
				url:  "https://www.google.com",
				exp:  1234,
			},
			verify: func(t *testing.T, repo Link, mockRedis *redis.MockRedis, args args) {
				url, err := repo.GetLink(ctx, args.code)
				require.NoError(t, err)
				require.Equal(t, args.url, url)
			},
		},
		{
			name: "should overwrite existing link successfully",
			args: args{
				code: "abc1234",
				url:  "https://www.google.com/v2",
				exp:  1234,
			},
			verify: func(t *testing.T, repo Link, mockRedis *redis.MockRedis, args args) {
				err := repo.SaveLink(ctx, args.code, "https://www.google.com/v1", 1234)
				require.NoError(t, err)
				err = repo.SaveLink(ctx, args.code, args.url, args.exp)
				require.NoError(t, err)
				url, err := repo.GetLink(ctx, args.code)
				require.NoError(t, err)
				require.Equal(t, args.url, url)
			},
		},
		{
			name: "shoud expire link after TTL exceeded",
			args: args{
				code: "abc1234",
				url:  "https://www.google.com",
				exp:  1,
			},
			verify: func(t *testing.T, repo Link, mockRedis *redis.MockRedis, args args) {
				exists, err := repo.CheckExists(ctx, args.code)
				require.NoError(t, err)
				assert.True(t, exists)

				mockRedis.Server.FastForward(2 * time.Second)

				exists, err = repo.CheckExists(
					ctx,
					args.code,
				)

				require.NoError(t, err)
				assert.False(t, exists)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo, mockRedis := newTestRepository(t)

			err := repo.SaveLink(ctx, tc.args.code, tc.args.url, tc.args.exp)

			require.NoError(t, err)

			tc.verify(t, repo, mockRedis, tc.args)
		})
	}
}

func TestLinkRepository_CheckExists(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	type args struct {
		code string
	}

	testCases := []struct {
		name          string
		setupDataFunc func(context.Context, Link)
		args          args
		verify        func(*testing.T, bool, error)
	}{
		{
			name: "should return true if code exists",
			setupDataFunc: func(_ context.Context, repo Link) {
				err := repo.SaveLink(ctx, "abc1234", "https://www.google.com", 1234)
				require.NoError(t, err)
			},
			args: args{
				code: "abc1234",
			},
			verify: func(t *testing.T, exists bool, err error) {
				require.NoError(t, err)
				assert.True(t, exists)
			},
		},
		{
			name: "should return false if code does not exist",
			setupDataFunc: func(_ context.Context, repo Link) {
			},
			args: args{
				code: "missing",
			},
			verify: func(t *testing.T, exists bool, err error) {
				require.NoError(t, err)
				assert.False(t, exists)
			},
		}, {
			name: "should return false if key expired",
			setupDataFunc: func(_ context.Context, repo Link) {
				err := repo.SaveLink(ctx, "abc1234", "https://www.google.com", 1)
				require.NoError(t, err)
				time.Sleep(2 * time.Second)
			},
			args: args{
				code: "def5678",
			},
			verify: func(t *testing.T, exists bool, err error) {
				require.NoError(t, err)
				assert.False(t, exists)
			},
		}, {
			name: "should return error if Redis client is unavailable",
			setupDataFunc: func(_ context.Context, repo Link) {
				err := repo.SaveLink(ctx, "abc1234", "https://www.google.com", 1234)
				require.NoError(t, err)
			},
			args: args{
				code: "abc1234",
			},
			verify: func(t *testing.T, exists bool, err error) {
				assert.Error(t, err)
				assert.False(t, exists)
				assert.Contains(t, err.Error(), "redis: client is closed")
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo, mockRedis := newTestRepository(t)

			tc.setupDataFunc(ctx, repo)

			if tc.name == "should return error if Redis client is unavailable" {
				mockRedis.Close()
			}

			exists, err := repo.CheckExists(ctx, tc.args.code)

			tc.verify(t, exists, err)
		})
	}
}
