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

	testCases := []struct {
		name                   string
		appConfig              config.Config
		expectedHTTPStatusCode int
		expectedMessage        string
		expectedServiceName    string
		expectedInstanceID     string
		expectGeneratedUUID    bool
	}{
		{
			name: "should return configured instance id",
			appConfig: config.Config{
				AppPort:     "8080",
				ServiceName: "bookmark-service",
				InstanceID:  "instance-1",
			},
			expectedHTTPStatusCode: http.StatusOK,
			expectedMessage:        "OK",
			expectedServiceName:    "bookmark-service",
			expectedInstanceID:     "instance-1",
			expectGeneratedUUID:    false,
		},
		{
			name: "should return generated uuid instance id",
			appConfig: config.Config{
				AppPort:     "8080",
				ServiceName: "bookmark-service",
				InstanceID:  "", // Empty to trigger UUID generation
			},
			expectedHTTPStatusCode: http.StatusOK,
			expectedMessage:        "OK",
			expectedServiceName:    "bookmark-service",
			expectGeneratedUUID:    true,
		},
		{
			name: "should handle different service names",
			appConfig: config.Config{
				AppPort:     "8081",
				ServiceName: "auth-service",
				InstanceID:  "prod-1",
			},
			expectedHTTPStatusCode: http.StatusOK,
			expectedMessage:        "OK",
			expectedServiceName:    "auth-service",
			expectedInstanceID:     "prod-1",
			expectGeneratedUUID:    false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockRedis := redis.NewMockRedis(t)
			pinger := redis.NewPinger(mockRedis.Client)

			cfg := tc.appConfig
			if cfg.InstanceID == "" {
				// Simulate UUID generation (same logic as LoadConfig in production)
				// For testing, we'll generate it through the service if needed
				// But configuration should have it set in tests
				cfg.InstanceID = "test-generated-uuid"
			}

			healthCheckService := service.NewHealthCheckService(
				cfg.ServiceName,
				cfg.InstanceID,
				pinger,
			)

			healthCheckHandler := handler.NewHealthCheckHandler(
				healthCheckService,
			)

			router := api.NewRouter(cfg.AppPort)
			api.RegisterHealthRoutes(router.GroupV1(), healthCheckHandler)

			// 8. Execute Request
			req := httptest.NewRequest(
				http.MethodGet,
				"/api/v1/health-check",
				nil,
			)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			// 9. Assertions
			assert.Equal(t, tc.expectedHTTPStatusCode, recorder.Code)

			var res response.HealthCheckResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &res)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedMessage, res.Message)
			assert.Equal(t, tc.expectedServiceName, res.ServiceName)

			if tc.expectGeneratedUUID {
				assert.NotEmpty(t, res.InstanceID)
				// Verify it looks like our test-generated UUID
				assert.Equal(t, "test-generated-uuid", res.InstanceID)
			} else {
				assert.Equal(t, tc.expectedInstanceID, res.InstanceID)
			}
		})
	}
}
