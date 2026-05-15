package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/huypham67/bookmark-management/internal/api"
	"github.com/huypham67/bookmark-management/internal/config"
	"github.com/huypham67/bookmark-management/internal/dto/response"
	"github.com/huypham67/bookmark-management/internal/handler"
	"github.com/huypham67/bookmark-management/internal/service"
	"github.com/huypham67/bookmark-management/pkg/redis"
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
		setupRedis func(*redis.MockRedis)
		expected   expected
	}{
		{
			name: "should return 200 OK with successful health check",
			appConfig: config.Config{
				AppPort:     "8080",
				ServiceName: "bookmark-service",
				InstanceID:  "instance-1",
			},
			setupRedis: func(mockRedis *redis.MockRedis) {
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
				AppPort:     "8080",
				ServiceName: "bookmark-service",
				InstanceID:  "instance-2",
			},
			setupRedis: func(mockRedis *redis.MockRedis) {
				mockRedis.Close()
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

			mockRedis := redis.NewMockRedis(t)
			tc.setupRedis(mockRedis)
			pinger := redis.NewPinger(mockRedis.Client)
			healthCheckService := service.NewHealthCheckService(tc.appConfig.ServiceName, tc.appConfig.InstanceID, pinger)
			healthCheckHandler := handler.NewHealthCheckHandler(healthCheckService)

			router := api.NewRouter()
			api.RegisterHealthRoutes(router.GroupV1(), healthCheckHandler)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/health-check", nil)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			assert.Equal(t, tc.expected.statusCode, recorder.Code)
			assert.Equal(t, "application/json; charset=utf-8", recorder.Header().Get("Content-Type"))

			var actual response.HealthCheckResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &actual)
			require.NoError(t, err)

			assert.Equal(t, tc.expected.response, actual)
		})
	}
}
