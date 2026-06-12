package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	bookmarkDTO "github.com/huypham67/bookmark-service/internal/dto/bookmark"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListBookmarksEndpoint(t *testing.T) {
	t.Parallel()

	type expected struct {
		statusCode    int
		bodyContains  string
		expectedCount int
		hasMoreData   bool
	}

	testCases := []struct {
		name        string
		queryParams string
		setupAuth   func(t *testing.T, app *AuthenticatedTestApp, req *http.Request)
		setupCache  func(t *testing.T, app *AuthenticatedTestApp)
		verifyCache func(t *testing.T, app *AuthenticatedTestApp)
		expected    expected
	}{
		{
			name:        "should return 401 when authorization header is missing",
			queryParams: "?page=1&limit=10",
			setupAuth: func(t *testing.T, app *AuthenticatedTestApp, req *http.Request) {
				// No auth header set
			},
			expected: expected{
				statusCode:   http.StatusUnauthorized,
				bodyContains: "missing authorization header",
			},
		},
		{
			name:        "should return 200 with empty bookmarks list when user has no bookmarks",
			queryParams: "?page=10&limit=10",
			setupAuth: func(t *testing.T, app *AuthenticatedTestApp, req *http.Request) {
				token, err := app.TokenGenerator.GenerateToken(
					"user-uuid-1",
					"testuser1",
					"testuser1@gmail.com",
				)
				require.NoError(t, err)
				req.Header.Set("Authorization", "Bearer "+token)
			},
			expected: expected{
				statusCode:    http.StatusOK,
				bodyContains:  "Bookmarks retrieved successfully!",
				expectedCount: 0,
				hasMoreData:   false,
			},
		},
		{
			name:        "should return bookmarks from cache when cache exists",
			queryParams: "?page=1&limit=10",
			setupCache: func(t *testing.T, app *AuthenticatedTestApp) {
				seedBookmarkListCache(t, app, 1, 10, "created_at")
			},
			setupAuth: func(t *testing.T, app *AuthenticatedTestApp, req *http.Request) {
				token, err := app.TokenGenerator.GenerateToken(
					"user-uuid-1",
					"testuser1",
					"testuser1@gmail.com",
				)
				require.NoError(t, err)
				req.Header.Set("Authorization", "Bearer "+token)
			},
			expected: expected{
				statusCode:    http.StatusOK,
				bodyContains:  "Bookmarks retrieved successfully!",
				expectedCount: 3, // cache holds 3 entries; the DB holds 8
				hasMoreData:   false,
			},
		},
		{
			name:        "should load bookmarks from database and populate cache when cache is empty",
			queryParams: "?page=1&limit=10",
			setupAuth: func(t *testing.T, app *AuthenticatedTestApp, req *http.Request) {
				token, err := app.TokenGenerator.GenerateToken(
					"user-uuid-1",
					"testuser1",
					"testuser1@gmail.com",
				)
				require.NoError(t, err)
				req.Header.Set("Authorization", "Bearer "+token)
			},
			verifyCache: func(t *testing.T, app *AuthenticatedTestApp) {
				hashKey := bookmarkCacheHashKey(cacheSeedUserID)
				require.True(t, app.MockRedis.Server.Exists(hashKey))
				assert.NotEmpty(t, app.MockRedis.Server.HGet(hashKey, bookmarkCacheFieldKey(1, 10, "created_at")))
			},
			expected: expected{
				statusCode:    http.StatusOK,
				bodyContains:  "Bookmarks retrieved successfully!",
				expectedCount: 8, // served from the DB on a cache miss
				hasMoreData:   false,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := setupBookmarkTestApp(t)

			if tc.setupCache != nil {
				tc.setupCache(t, app)
			}

			httpRequest := httptest.NewRequest(
				http.MethodGet,
				"/api/bookmark_service/v1/bookmarks"+tc.queryParams,
				nil,
			)

			tc.setupAuth(t, app, httpRequest)

			httpRecorder := httptest.NewRecorder()

			app.Router.ServeHTTP(httpRecorder, httpRequest)

			assert.Equal(t, tc.expected.statusCode, httpRecorder.Code)
			assert.Contains(t, httpRecorder.Body.String(), tc.expected.bodyContains)

			if tc.expected.statusCode == http.StatusOK {
				var resp bookmarkDTO.BookmarkListResponse
				err := json.Unmarshal(httpRecorder.Body.Bytes(), &resp)
				require.NoError(t, err)

				assert.Equal(t, tc.expected.expectedCount, len(resp.Data))
				assert.NotEmpty(t, resp.Pagination)
				assert.Equal(t, int64(10), resp.Pagination.Limit)
			}

			if tc.verifyCache != nil {
				tc.verifyCache(t, app)
			}
		})
	}
}
