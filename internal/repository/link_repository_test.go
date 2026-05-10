package repository

import (
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/huypham67/bookmark-management/infrastructure/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setUpMockRedis creates a miniredis instance and returns a Redis client for testing
func setUpMockRedis(t *testing.T) (*redis.RedisClient, func()) {
	mr := miniredis.NewMiniRedis()
	err := mr.Start()
	require.NoError(t, err)

	client, err := redis.NewRedisClient(redis.RedisConfig{
		Host:     "localhost",
		Port:     fmt.Sprintf("%d", mr.Server().Addr().Port),
		Password: "",
		Database: 0,
	})
	require.NoError(t, err)

	return client, func() {
		mr.Close()
	}
}

func TestLinkRepository_SaveLink(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		code      string
		url       string
		exp       int64
		expectErr bool
	}{
		{
			name:      "should save link successfully with valid parameters",
			code:      "abc123",
			url:       "https://example.com",
			exp:       3600,
			expectErr: false,
		},
		{
			name:      "should save link with short expiration time",
			code:      "xyz789",
			url:       "https://google.com",
			exp:       60,
			expectErr: false,
		},
		{
			name:      "should save link with zero expiration (no expiration)",
			code:      "noexp01",
			url:       "https://github.com",
			exp:       0,
			expectErr: false,
		},
		{
			name:      "should save link with very long URL",
			code:      "longurl",
			url:       "https://example.com/very/long/path?param1=value1&param2=value2&param3=value3&param4=value4",
			exp:       7200,
			expectErr: false,
		},
		{
			name:      "should save link with empty code",
			code:      "",
			url:       "https://example.com",
			exp:       3600,
			expectErr: false,
		},
		{
			name:      "should save link with empty URL",
			code:      "empty01",
			url:       "",
			exp:       3600,
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			redisClient, cleanup := setUpMockRedis(t)
			defer cleanup()

			repo := NewLinkRepository(redisClient)

			err := repo.SaveLink(tc.code, tc.url, tc.exp)

			if tc.expectErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestLinkRepository_CheckExists(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		setupData    func(*redis.RedisClient)
		code         string
		expectExists bool
		expectErr    bool
	}{
		{
			name: "should return true when code exists",
			setupData: func(client *redis.RedisClient) {
				err := client.Set("abc123", "https://example.com", 3600*time.Second)
				require.NoError(t, err)
			},
			code:         "abc123",
			expectExists: true,
			expectErr:    false,
		},
		{
			name:         "should return false when code does not exist",
			setupData:    func(client *redis.RedisClient) {},
			code:         "nonexistent",
			expectExists: false,
			expectErr:    false,
		},
		{
			name: "should return true when multiple codes exist and checking middle one",
			setupData: func(client *redis.RedisClient) {
				err := client.Set("code1", "https://example1.com", 3600*time.Second)
				require.NoError(t, err)
				err = client.Set("code2", "https://example2.com", 3600*time.Second)
				require.NoError(t, err)
				err = client.Set("code3", "https://example3.com", 3600*time.Second)
				require.NoError(t, err)
			},
			code:         "code2",
			expectExists: true,
			expectErr:    false,
		},
		{
			name:         "should return false when checking empty code",
			setupData:    func(client *redis.RedisClient) {},
			code:         "",
			expectExists: false,
			expectErr:    false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			redisClient, cleanup := setUpMockRedis(t)
			defer cleanup()

			tc.setupData(redisClient)

			repo := NewLinkRepository(redisClient)

			exists, err := repo.CheckExists(tc.code)

			if tc.expectErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.expectExists, exists)
		})
	}
}

func TestLinkRepository_SaveAndCheckExists(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		code      string
		url       string
		exp       int64
		expectErr bool
	}{
		{
			name:      "should save link and verify it exists",
			code:      "verify01",
			url:       "https://example.com",
			exp:       3600,
			expectErr: false,
		},
		{
			name:      "should save multiple links and verify they exist",
			code:      "multi01",
			url:       "https://example.com/path1",
			exp:       7200,
			expectErr: false,
		},
		{
			name:      "should save link with special characters in URL",
			code:      "special",
			url:       "https://example.com?q=test&sort=asc#section",
			exp:       3600,
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			redisClient, cleanup := setUpMockRedis(t)
			defer cleanup()

			repo := NewLinkRepository(redisClient)

			// Save the link
			err := repo.SaveLink(tc.code, tc.url, tc.exp)
			assert.NoError(t, err)

			// Verify it exists
			exists, err := repo.CheckExists(tc.code)
			assert.NoError(t, err)
			assert.True(t, exists, "saved link should exist")
		})
	}
}

func TestLinkRepository_OverwriteExistingLink(t *testing.T) {
	t.Parallel()

	redisClient, cleanup := setUpMockRedis(t)
	defer cleanup()

	repo := NewLinkRepository(redisClient)

	code := "overwrite01"
	firstURL := "https://example.com/api/v1"
	secondURL := "https://example.com/api/v2"

	// Save first link
	err := repo.SaveLink(code, firstURL, 3600)
	assert.NoError(t, err)

	exists, err := repo.CheckExists(code)
	assert.NoError(t, err)
	assert.True(t, exists)

	// Overwrite with second link
	err = repo.SaveLink(code, secondURL, 3600)
	assert.NoError(t, err)

	exists, err = repo.CheckExists(code)
	assert.NoError(t, err)
	assert.True(t, exists, "overwritten link should still exist")
}
