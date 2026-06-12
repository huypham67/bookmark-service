package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	bookmarkDTO "github.com/huypham67/bookmark-service/internal/dto/bookmark"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateBookmarkEndpoint(t *testing.T) {
	t.Parallel()

	type expected struct {
		statusCode   int
		bodyContains string
	}

	testCases := []struct {
		name        string
		requestBody string
		setupAuth   func(t *testing.T, app *AuthenticatedTestApp, req *http.Request)
		setupCache  func(t *testing.T, app *AuthenticatedTestApp)
		verifyCache func(t *testing.T, app *AuthenticatedTestApp)
		expected    expected
	}{
		{
			name: "should return 401 when the Authorization header is missing",
			requestBody: `{
				"description": "Test Bookmark",
				"url": "https://test-example.com"
			}`,
			setupAuth: func(t *testing.T, app *AuthenticatedTestApp, req *http.Request) {
				// No auth header set
			},
			expected: expected{
				statusCode:   http.StatusUnauthorized,
				bodyContains: "missing authorization header",
			},
		},
		{
			name:        "should return 400 when the request body is malformed JSON",
			requestBody: `{invalid json}`,
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
				statusCode:   http.StatusBadRequest,
				bodyContains: "Invalid request body",
			},
		},
		{
			name: "should return 201, persist the bookmark, and invalidate the user's list cache",
			requestBody: `{
				"description": "New Bookmark",
				"url": "https://new-example.com"
			}`,
			setupCache: func(t *testing.T, app *AuthenticatedTestApp) {
				seedBookmarkListCache(t, app, 1, 10, "created_at")
				require.True(t, app.MockRedis.Server.Exists(bookmarkCacheHashKey(cacheSeedUserID)))
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
			verifyCache: func(t *testing.T, app *AuthenticatedTestApp) {
				assert.False(t, app.MockRedis.Server.Exists(bookmarkCacheHashKey(cacheSeedUserID)))
			},
			expected: expected{
				statusCode:   http.StatusCreated,
				bodyContains: "Bookmark created successfully!",
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
				http.MethodPost,
				"/api/bookmark_service/v1/bookmarks",
				bytes.NewBufferString(tc.requestBody),
			)

			httpRequest.Header.Set("Content-Type", "application/json")

			tc.setupAuth(t, app, httpRequest)

			httpRecorder := httptest.NewRecorder()

			app.Router.ServeHTTP(httpRecorder, httpRequest)

			assert.Equal(t, tc.expected.statusCode, httpRecorder.Code)
			assert.Contains(t, httpRecorder.Body.String(), tc.expected.bodyContains)

			if tc.expected.statusCode == http.StatusCreated {
				var resp bookmarkDTO.BookmarkResponse
				err := json.Unmarshal(httpRecorder.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.NotEmpty(t, resp.Data.ID)
				assert.NotEmpty(t, resp.Data.Code)
				assert.NotEmpty(t, resp.Data.Description)
				assert.NotEmpty(t, resp.Data.URL)
			}

			if tc.verifyCache != nil {
				tc.verifyCache(t, app)
			}
		})
	}
}
