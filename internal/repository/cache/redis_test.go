package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepository_SetCache(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		wantErr bool
		verify  func(*testing.T, Repository, error)
	}{
		{
			name:    "should set cache successfully",
			wantErr: false,
			verify: func(t *testing.T, repo Repository, err error) {
				require.NoError(t, err)
				ctx := context.Background()
				result, err := repo.GetCache(ctx, "bookmarks", "user:123")
				require.NoError(t, err)
				assert.Equal(t, []byte(`{"id":"123"}`), result)
			},
		},
		{
			name:    "should return error if Redis client is closed",
			wantErr: true,
			verify: func(t *testing.T, _ Repository, err error) {
				assert.Error(t, err)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			repo, mockRedis := newTestRepository(t)

			if tc.name == "should return error if Redis client is closed" {
				mockRedis.Close()
			}

			err := repo.SetCache(ctx, "bookmarks", "user:123", []byte(`{"id":"123"}`), 1*time.Hour)

			tc.verify(t, repo, err)
		})
	}
}

func TestRepository_GetCache(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		setupDataFunc func(context.Context, Repository)
		verify        func(*testing.T, []byte, error)
	}{
		{
			name: "should return value when cache exists",
			setupDataFunc: func(ctx context.Context, repo Repository) {
				_ = repo.SetCache(ctx, "bookmarks", "user:123", []byte(`{"id":"123"}`), 1*time.Hour)
			},
			verify: func(t *testing.T, result []byte, err error) {
				require.NoError(t, err)
				assert.Equal(t, []byte(`{"id":"123"}`), result)
			},
		},
		{
			name:          "should return ErrCacheMiss when cache does not exist",
			setupDataFunc: func(_ context.Context, _ Repository) {},
			verify: func(t *testing.T, result []byte, err error) {
				require.Error(t, err)
				assert.ErrorIs(t, err, ErrCacheMiss)
			},
		},
		{
			name:          "should return error if Redis client is unavailable",
			setupDataFunc: func(ctx context.Context, repo Repository) {},
			verify: func(t *testing.T, result []byte, err error) {
				assert.Error(t, err)
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

			result, err := repo.GetCache(ctx, "bookmarks", "user:123")

			tc.verify(t, result, err)
		})
	}
}

func TestRepository_DeleteCacheByFieldKey(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		setupDataFunc func(context.Context, Repository)
		verify        func(*testing.T, Repository, error)
	}{
		{
			name: "should delete field successfully",
			setupDataFunc: func(ctx context.Context, repo Repository) {
				_ = repo.SetCache(ctx, "bookmarks", "user:123", []byte("test"), 1*time.Hour)
			},
			verify: func(t *testing.T, repo Repository, err error) {
				require.NoError(t, err)
				ctx := context.Background()
				_, getErr := repo.GetCache(ctx, "bookmarks", "user:123")
				assert.ErrorIs(t, getErr, ErrCacheMiss)
			},
		},
		{
			name:          "should return error if Redis client is unavailable",
			setupDataFunc: func(_ context.Context, _ Repository) {},
			verify: func(t *testing.T, _ Repository, err error) {
				assert.Error(t, err)
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

			err := repo.DeleteCacheByFieldKey(ctx, "bookmarks", "user:123")

			tc.verify(t, repo, err)
		})
	}
}

func TestRepository_DeleteCacheByHashKey(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		setupDataFunc func(context.Context, Repository)
		verify        func(*testing.T, Repository, error)
	}{
		{
			name: "should delete entire hash successfully",
			setupDataFunc: func(ctx context.Context, repo Repository) {
				_ = repo.SetCache(ctx, "bookmarks", "user:123", []byte("test1"), 1*time.Hour)
				_ = repo.SetCache(ctx, "bookmarks", "user:456", []byte("test2"), 1*time.Hour)
			},
			verify: func(t *testing.T, repo Repository, err error) {
				require.NoError(t, err)
				ctx := context.Background()
				_, err1 := repo.GetCache(ctx, "bookmarks", "user:123")
				assert.ErrorIs(t, err1, ErrCacheMiss)
				_, err2 := repo.GetCache(ctx, "bookmarks", "user:456")
				assert.ErrorIs(t, err2, ErrCacheMiss)
			},
		},
		{
			name:          "should return error if Redis client is unavailable",
			setupDataFunc: func(_ context.Context, _ Repository) {},
			verify: func(t *testing.T, _ Repository, err error) {
				assert.Error(t, err)
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

			err := repo.DeleteCacheByHashKey(ctx, "bookmarks")

			tc.verify(t, repo, err)
		})
	}
}
