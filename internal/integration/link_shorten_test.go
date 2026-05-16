package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/huypham67/bookmark-service/internal/dto/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShortenURLEndpoint(t *testing.T) {
	t.Parallel()

	type expected struct {
		statusCode   int
		bodyContains string
	}

	testCases := []struct {
		name        string
		requestBody string
		setupRedis  func(*TestApp)
		expected    expected
	}{
		{
			name: "should return 200 when request is valid",
			requestBody: `
			{
				"url": "https://example.com",
				"exp": 3600
			}
			`,
			setupRedis: func(app *TestApp) {
			},
			expected: expected{
				statusCode:   http.StatusOK,
				bodyContains: "Shorten URL generated successfully",
			},
		},
		{
			name:        "should return 400 when request body is invalid JSON",
			requestBody: `{invalid json}`,
			setupRedis: func(app *TestApp) {
			},
			expected: expected{
				statusCode:   http.StatusBadRequest,
				bodyContains: "Invalid request body",
			},
		},
		{
			name: "should return 400 when validation fails",
			requestBody: `
			{
				"url": "",
				"exp": 3600
			}
			`,
			setupRedis: func(app *TestApp) {
			},
			expected: expected{
				statusCode:   http.StatusBadRequest,
				bodyContains: "Invalid request body",
			},
		},
		{
			name: "should return 500 when redis connection fails",
			requestBody: `
			{
				"url": "https://example.com",
				"exp": 3600
			}
			`,
			setupRedis: func(app *TestApp) {
				app.MockRedis.Close()
			},
			expected: expected{
				statusCode:   http.StatusInternalServerError,
				bodyContains: "Internal Server Error",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := setupLinkTestApp(t)
			tc.setupRedis(app)

			httpRequest := httptest.NewRequest(http.MethodPost, "/api/v1/links/shorten", bytes.NewBufferString(tc.requestBody))

			httpRequest.Header.Set("Content-Type", "application/json")
			httpRecorder := httptest.NewRecorder()
			app.Router.ServeHTTP(httpRecorder, httpRequest)

			assert.Equal(t, tc.expected.statusCode, httpRecorder.Code)
			assert.Equal(t, "application/json; charset=utf-8", httpRecorder.Header().Get("Content-Type"))
			assert.Contains(t, httpRecorder.Body.String(), tc.expected.bodyContains)

			if tc.expected.statusCode == http.StatusOK {
				var actual response.ShortenURLResponse

				err := json.Unmarshal(httpRecorder.Body.Bytes(), &actual)

				require.NoError(t, err)
				assert.NotEmpty(t, actual.Code)
			}
		})
	}
}
