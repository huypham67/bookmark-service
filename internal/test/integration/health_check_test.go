package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/huypham67/bookmark-service/internal/bootstrap"
	healthDTO "github.com/huypham67/bookmark-service/internal/dto/health"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthCheckEndpoint(t *testing.T) {
	t.Parallel()

	type expected struct {
		statusCode int
		response   healthDTO.HealthCheckResponse
	}

	testCases := []struct {
		name      string
		appConfig bootstrap.Config
		setupDeps func(*testing.T, *TestApp)
		expected  expected
	}{
		{
			name: "should return 200 OK with successful health check",
			appConfig: bootstrap.Config{
				ServiceName: "bookmark-service",
				InstanceID:  "instance-1",
			},
			setupDeps: func(t *testing.T, app *TestApp) {
			},
			expected: expected{
				statusCode: http.StatusOK,
				response: healthDTO.HealthCheckResponse{
					Message:     "OK",
					ServiceName: "bookmark-service",
					InstanceID:  "instance-1",
				},
			},
		},
		{
			name: "should return 500 when redis connection fails",
			appConfig: bootstrap.Config{
				ServiceName: "bookmark-service",
				InstanceID:  "instance-2",
			},
			setupDeps: func(t *testing.T, app *TestApp) {
				app.MockRedis.Close()
			},
			expected: expected{
				statusCode: http.StatusInternalServerError,
				response: healthDTO.HealthCheckResponse{
					Message:     "FAILED",
					ServiceName: "bookmark-service",
					InstanceID:  "instance-2",
				},
			},
		},
		{
			name: "should return 500 when database connection fails",
			appConfig: bootstrap.Config{
				ServiceName: "bookmark-service",
				InstanceID:  "instance-3",
			},
			setupDeps: func(t *testing.T, app *TestApp) {
				sqlDB, err := app.MockDB.DB()
				require.NoError(t, err)
				require.NoError(t, sqlDB.Close())
			},
			expected: expected{
				statusCode: http.StatusInternalServerError,
				response: healthDTO.HealthCheckResponse{
					Message:     "FAILED",
					ServiceName: "bookmark-service",
					InstanceID:  "instance-3",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := setupHealthCheckTestApp(t, tc.appConfig.ServiceName, tc.appConfig.InstanceID)

			tc.setupDeps(t, app)

			req := httptest.NewRequest(http.MethodGet, "/api/bookmark_service/health-check", nil)
			recorder := httptest.NewRecorder()
			app.Router.ServeHTTP(recorder, req)

			assert.Equal(t, tc.expected.statusCode, recorder.Code)
			assert.Equal(t, "application/json; charset=utf-8", recorder.Header().Get("Content-Type"))

			var actual healthDTO.HealthCheckResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &actual)
			require.NoError(t, err)

			assert.Equal(t, tc.expected.response, actual)
		})
	}
}
