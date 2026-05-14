package repository

import (
	"context"
	"strings"
	"testing"
	"time"

	pkgRedis "github.com/huypham67/bookmark-management/pkg/redis"
	redisClient "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRepository(t *testing.T) Link {
	t.Helper()

	mockRedis := pkgRedis.NewMockRedis(t)

	client := &pkgRedis.RedisClient{
		Client: mockRedis.Client,
	}

	return NewLinkRepository(client)
}

func TestLinkRepository_SaveLink(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		code string
		url  string
		exp  int64
	}{
		{
			name: "should save link successfully",
			code: "abc123",
			url:  "https://example.com",
			exp:  3600,
		},
		{
			name: "should save link with zero expiration",
			code: "no-exp",
			url:  "https://github.com",
			exp:  0,
		},
		{
			name: "should save link with empty url",
			code: "empty-url",
			url:  "",
			exp:  3600,
		},
		{
			name: "should save link with long url",
			code: "long-url",
			url:  "https://example.com/" + strings.Repeat("a", 4000),
			exp:  7200,
		},
		{
			name: "should save special character url",
			code: "special-url",
			url:  "https://example.com?q=hello world&emoji=🔥",
			exp:  3600,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := newTestRepository(t)

			err := repo.SaveLink(tc.code, tc.url, tc.exp)

			require.NoError(t, err)

			exists, err := repo.CheckExists(tc.code)

			require.NoError(t, err)
			assert.True(t, exists)

			value, err := repo.GetLink(tc.code)

			require.NoError(t, err)
			assert.Equal(t, tc.url, value)
		})
	}
}

func TestLinkRepository_CheckExists(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		setupData    func(t *testing.T, repo Link, mockRedis *pkgRedis.MockRedis)
		code         string
		expectExists bool
		expectError  bool
		useMockRedis bool
	}{
		{
			name: "should return true when code exists",
			setupData: func(t *testing.T, repo Link, mockRedis *pkgRedis.MockRedis) {
				err := repo.SaveLink(
					"abc123",
					"https://example.com",
					3600,
				)

				require.NoError(t, err)
			},
			code:         "abc123",
			expectExists: true,
			expectError:  false,
			useMockRedis: false,
		},
		{
			name:         "should return false when code does not exist",
			setupData:    func(t *testing.T, repo Link, mockRedis *pkgRedis.MockRedis) {},
			code:         "not-found",
			expectExists: false,
			expectError:  false,
			useMockRedis: false,
		},
		{
			name:         "should return false with empty code",
			setupData:    func(t *testing.T, repo Link, mockRedis *pkgRedis.MockRedis) {},
			code:         "",
			expectExists: false,
			expectError:  false,
			useMockRedis: false,
		},
		{
			name: "should return error when Redis connection fails",
			setupData: func(t *testing.T, repo Link, mockRedis *pkgRedis.MockRedis) {
				// Close Redis server to simulate connection error
				mockRedis.Server.Close()
			},
			code:         "test-code",
			expectExists: false,
			expectError:  true,
			useMockRedis: true,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var repo Link
			var mockRedis *pkgRedis.MockRedis

			if tc.useMockRedis {
				mockRedis = pkgRedis.NewMockRedis(t)
				client := &pkgRedis.RedisClient{
					Client: mockRedis.Client,
				}
				repo = NewLinkRepository(client)
			} else {
				repo = newTestRepository(t)
				mockRedis = pkgRedis.NewMockRedis(t)
			}

			tc.setupData(t, repo, mockRedis)

			exists, err := repo.CheckExists(tc.code)

			if tc.expectError {
				assert.Error(t, err)
				assert.False(t, exists)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectExists, exists)
			}
		})
	}
}

func TestLinkRepository_GetLink(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		setupData func(t *testing.T, repo Link)
		code      string
		expected  string
		expectErr bool
	}{
		{
			name: "should get existing link",
			setupData: func(t *testing.T, repo Link) {
				err := repo.SaveLink(
					"abc123",
					"https://example.com",
					3600,
				)

				require.NoError(t, err)
			},
			code:      "abc123",
			expected:  "https://example.com",
			expectErr: false,
		},
		{
			name:      "should return error when code not found",
			setupData: func(t *testing.T, repo Link) {},
			code:      "not-found",
			expected:  "",
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := newTestRepository(t)

			tc.setupData(t, repo)

			url, err := repo.GetLink(tc.code)

			if tc.expectErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.expected, url)
		})
	}
}

func TestLinkRepository_SaveAndGetLink(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)

	code := "integration-test"
	expectedURL := "https://example.com"

	err := repo.SaveLink(code, expectedURL, 3600)
	require.NoError(t, err)

	url, err := repo.GetLink(code)

	assert.NoError(t, err)
	assert.Equal(t, expectedURL, url)
}

func TestLinkRepository_LinkExpiration(t *testing.T) {
	t.Parallel()

	mockRedis := pkgRedis.NewMockRedis(t)

	client := &pkgRedis.RedisClient{
		Client: mockRedis.Client,
	}

	repo := NewLinkRepository(client)

	code := "exp-test"

	err := repo.SaveLink(
		code,
		"https://example.com",
		1,
	)

	require.NoError(t, err)

	exists, err := repo.CheckExists(code)

	require.NoError(t, err)
	assert.True(t, exists)

	mockRedis.Server.FastForward(2 * time.Second)

	exists, err = repo.CheckExists(code)

	require.NoError(t, err)
	assert.False(t, exists)
}

func TestLinkRepository_OverwriteExistingLink(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)

	code := "overwrite"

	err := repo.SaveLink(
		code,
		"https://example.com/v1",
		3600,
	)

	require.NoError(t, err)

	err = repo.SaveLink(
		code,
		"https://example.com/v2",
		3600,
	)

	require.NoError(t, err)

	url, err := repo.GetLink(code)

	assert.NoError(t, err)
	assert.Equal(t, "https://example.com/v2", url)
}

func TestLinkRepository_RedisFailure(t *testing.T) {
	t.Parallel()

	mockRedis := pkgRedis.NewMockRedis(t)

	client := &pkgRedis.RedisClient{
		Client: mockRedis.Client,
	}

	repo := NewLinkRepository(client)

	mockRedis.Server.Close()

	err := repo.SaveLink(
		"abc",
		"https://example.com",
		3600,
	)

	assert.Error(t, err)
}

func TestLinkRepository_RawRedisValidation(t *testing.T) {
	t.Parallel()

	mockRedis := pkgRedis.NewMockRedis(t)

	client := &pkgRedis.RedisClient{
		Client: mockRedis.Client,
	}

	repo := NewLinkRepository(client)

	code := "raw-check"
	expectedURL := "https://example.com"

	err := repo.SaveLink(code, expectedURL, 3600)

	require.NoError(t, err)

	rawURL, err := client.Client.Get(
		context.Background(),
		code,
	).Result()

	assert.NoError(t, err)
	assert.Equal(t, expectedURL, rawURL)
}

func TestLinkRepository_WithRealRedisCommand(t *testing.T) {
	t.Parallel()

	mockRedis := pkgRedis.NewMockRedis(t)

	redisCmd := mockRedis.Client.Ping(context.Background())

	assert.Equal(t, "PONG", redisCmd.Val())
	assert.NoError(t, redisCmd.Err())

	_ = redisClient.Nil
}
