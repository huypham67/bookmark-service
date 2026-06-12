package link

import (
	"context"
	"testing"
	"time"

	"github.com/huypham67/bookmark-common/pkg/dbutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepository_GetLink(t *testing.T) {
	t.Parallel()

	type args struct {
		code string
	}

	testCases := []struct {
		name          string
		setupDataFunc func(context.Context, Repository)
		args          args
		verify        func(*testing.T, string, error)
	}{
		{
			name: "should return URL when code exists",
			setupDataFunc: func(ctx context.Context, repo Repository) {
				err := repo.SaveLink(ctx, "abc1234", "https://www.google.com", 1234)
				require.NoError(t, err)
			},
			args: args{
				code: "abc1234",
			},
			verify: func(t *testing.T, url string, err error) {
				require.NoError(t, err)
				assert.Equal(t, "https://www.google.com", url)
			},
		},
		{
			name: "should return error when code does not exist",
			setupDataFunc: func(_ context.Context, _ Repository) {
			},
			args: args{
				code: "nonexistent",
			},
			verify: func(t *testing.T, url string, err error) {
				require.Error(t, err)
				assert.Empty(t, url)
				assert.ErrorIs(t, err, dbutils.ErrRecordNotFoundType)
			},
		},
		{
			name: "should return error if Redis client is unavailable",
			setupDataFunc: func(ctx context.Context, repo Repository) {
				err := repo.SaveLink(ctx, "abc1234", "https://www.google.com", 1234)
				require.NoError(t, err)
			},
			args: args{
				code: "abc1234",
			},
			verify: func(t *testing.T, url string, err error) {
				assert.Error(t, err)
				assert.Empty(t, url)
				assert.Contains(t, err.Error(), "redis: client is closed")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			repo, mockRedis := newTestRepository(t)

			tc.setupDataFunc(ctx, repo)

			if tc.name == "should return error if Redis client is unavailable" {
				mockRedis.Close()
			}

			url, err := repo.GetLink(ctx, tc.args.code)

			tc.verify(t, url, err)
		})
	}
}

func TestRepository_CheckExists(t *testing.T) {
	t.Parallel()

	type args struct {
		code string
	}

	testCases := []struct {
		name          string
		setupDataFunc func(context.Context, Repository)
		args          args
		verify        func(*testing.T, bool, error)
	}{
		{
			name: "should return true if code exists",
			setupDataFunc: func(ctx context.Context, repo Repository) {
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
			setupDataFunc: func(_ context.Context, _ Repository) {
			},
			args: args{
				code: "missing",
			},
			verify: func(t *testing.T, exists bool, err error) {
				require.NoError(t, err)
				assert.False(t, exists)
			},
		},
		{
			name: "should return false if key expired",
			setupDataFunc: func(ctx context.Context, repo Repository) {
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
		},
		{
			name: "should return error if Redis client is unavailable",
			setupDataFunc: func(ctx context.Context, repo Repository) {
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
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

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
