package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedirectToURLEndpoint(t *testing.T) {
	t.Parallel()

	type expected struct {
		statusCode   int
		location     string
		bodyContains string
	}

	testCases := []struct {
		name       string
		code       string
		setupRedis func(*TestApp)
		expected   expected
	}{
		{
			name: "should redirect successfully when code exists",
			code: "abc1234",
			setupRedis: func(app *TestApp) {
				app.MockRedis.Server.Set("abc1234", "https://www.google.com")
			},
			expected: expected{
				statusCode: http.StatusFound,
				location:   "https://www.google.com",
			},
		},
		{
			name: "should return 404 when code does not exist",
			code: "missing",
			setupRedis: func(app *TestApp) {
				// No setup needed since the code does not exist
			},
			expected: expected{
				statusCode:   http.StatusNotFound,
				bodyContains: "Short link not found",
			},
		},
		{
			name: "should return 404 when redis connection fails",
			code: "abc1234",
			setupRedis: func(app *TestApp) {
				app.MockRedis.Close()
			},
			expected: expected{
				statusCode:   http.StatusNotFound,
				bodyContains: "Short link not found",
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := setupLinkTestApp(t)

			tc.setupRedis(app)

			httpRequest := httptest.NewRequest(
				http.MethodGet,
				"/api/v1/links/redirect/"+tc.code,
				nil,
			)

			httpRecorder := httptest.NewRecorder()

			app.Router.ServeHTTP(httpRecorder, httpRequest)

			assert.Equal(
				t,
				tc.expected.statusCode,
				httpRecorder.Code,
			)

			if tc.expected.location != "" {
				assert.Equal(
					t,
					tc.expected.location,
					httpRecorder.Header().Get("Location"),
				)

				return
			}

			assert.Equal(
				t,
				"application/json; charset=utf-8",
				httpRecorder.Header().Get("Content-Type"),
			)

			assert.Contains(
				t,
				httpRecorder.Body.String(),
				tc.expected.bodyContains,
			)
		})
	}
}
