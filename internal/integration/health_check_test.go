package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/huypham67/bookmark-management/infrastructure/redis"
	"github.com/huypham67/bookmark-management/internal/api"
	"github.com/huypham67/bookmark-management/internal/config"
	"github.com/huypham67/bookmark-management/internal/dto/response"
	healthHandler "github.com/huypham67/bookmark-management/internal/handler/health"
	healthService "github.com/huypham67/bookmark-management/internal/service/health"
	"github.com/huypham67/bookmark-management/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthCheckEndpoint(t *testing.T) {

	testCases := []struct {
		name                   string
		appConfig              config.AppConfig
		expectedHTTPStatusCode int
		expectedMessage        string
		expectedServiceName    string
		expectedInstanceID     string
		expectGeneratedUUID    bool
	}{
		{
			name: "should return configured instance id",

			appConfig: config.AppConfig{
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

			appConfig: config.AppConfig{
				AppPort:     "8080",
				ServiceName: "bookmark-service",
			},

			expectedHTTPStatusCode: http.StatusOK,
			expectedMessage:        "OK",
			expectedServiceName:    "bookmark-service",
			expectGeneratedUUID:    true,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {

			t.Setenv("APP_PORT", testCase.appConfig.AppPort)
			t.Setenv("SERVICE_NAME", testCase.appConfig.ServiceName)
			if testCase.appConfig.InstanceID != "" {
				t.Setenv(
					"INSTANCE_ID",
					testCase.appConfig.InstanceID,
				)
			} else {
				_ = os.Unsetenv("INSTANCE_ID")
			}

			cfg, err := config.LoadConfig()

			require.NoError(t, err)

			// Set up miniredis for testing
			mr := miniredis.NewMiniRedis()
			require.NoError(t, mr.Start())
			defer mr.Close()

			redisClient, err := redis.NewRedisClient(redis.RedisConfig{
				Host:     "localhost",
				Port:     fmt.Sprintf("%d", mr.Server().Addr().Port),
				Password: "",
				Database: 0,
			})
			require.NoError(t, err)

			healthCheckService := healthService.NewHealthCheckService(
				cfg.ServiceName,
				cfg.InstanceID,
				redisClient,
			)

			healthCheckHandler := healthHandler.NewHealthCheckHandler(
				healthCheckService,
			)

			// Create a mock linkHandler
			mockLinkHandler := mocks.NewLinkHandler(t)

			router := api.NewRouter(
				cfg.AppPort,
				healthCheckHandler,
				mockLinkHandler,
			)

			req := httptest.NewRequest(
				http.MethodGet,
				"/api/v1/health-check",
				nil,
			)

			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, req)

			require.Equal(
				t,
				testCase.expectedHTTPStatusCode,
				recorder.Code,
			)

			var res response.HealthCheckResponse

			err = json.Unmarshal(
				recorder.Body.Bytes(),
				&res,
			)

			require.NoError(t, err)

			assert.Equal(
				t,
				testCase.expectedMessage,
				res.Message,
			)

			assert.Equal(
				t,
				testCase.expectedServiceName,
				res.ServiceName,
			)

			if testCase.expectGeneratedUUID {
				assert.NotEmpty(
					t,
					res.InstanceID,
				)
			} else {
				assert.Equal(
					t,
					testCase.expectedInstanceID,
					res.InstanceID,
				)
			}
		})
	}
}
