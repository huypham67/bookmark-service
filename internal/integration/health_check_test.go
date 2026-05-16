package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/huypham67/bookmark-service/internal/config"
	"github.com/huypham67/bookmark-service/internal/dto/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthCheckEndpoint(t *testing.T) {
	t.Parallel()

	type expected struct {
		statusCode int
		response   response.HealthCheckResponse
	}

	testCases := []struct {
		name       string
		appConfig  config.Config
		setupRedis func(*TestApp)
		expected   expected
	}{
		{
			name: "should return 200 OK with successful health check",
			appConfig: config.Config{
				ServiceName: "bookmark-service",
				InstanceID:  "instance-1",
			},
			setupRedis: func(app *TestApp) {
			},
			expected: expected{
				statusCode: http.StatusOK,
				response: response.HealthCheckResponse{
					Message:     "OK",
					ServiceName: "bookmark-service",
					InstanceID:  "instance-1",
				},
			},
		},
		{
			name: "should return 500 when redis connection fails",
			appConfig: config.Config{
				ServiceName: "bookmark-service",
				InstanceID:  "instance-2",
			},
			setupRedis: func(app *TestApp) {
				app.MockRedis.Close()
			},
			expected: expected{
				statusCode: http.StatusInternalServerError,
				response: response.HealthCheckResponse{
					Message:     "FAILED",
					ServiceName: "bookmark-service",
					InstanceID:  "instance-2",
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := setupHealthCheckTestApp(t, tc.appConfig.ServiceName, tc.appConfig.InstanceID)

			tc.setupRedis(app)

			req := httptest.NewRequest(http.MethodGet, "/api/health-check", nil)
			recorder := httptest.NewRecorder()
			app.Router.ServeHTTP(recorder, req)

			assert.Equal(t, tc.expected.statusCode, recorder.Code)
			assert.Equal(t, "application/json; charset=utf-8", recorder.Header().Get("Content-Type"))

			var actual response.HealthCheckResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &actual)
			require.NoError(t, err)

			assert.Equal(t, tc.expected.response, actual)
		})
	}
}
