package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/huypham67/bookmark-management/internal/api"
	"github.com/huypham67/bookmark-management/internal/config"
	"github.com/huypham67/bookmark-management/internal/handler"
	"github.com/huypham67/bookmark-management/internal/model"
	"github.com/huypham67/bookmark-management/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthCheckEndpoint(t *testing.T) {

	testCases := []struct {
		name                   string
		config                 config.Config
		expectedHTTPStatusCode int
		expectedMessage        string
		expectedServiceName    string
		expectedInstanceID     string
		expectGeneratedUUID    bool
	}{
		{
			name: "should return configured instance id",

			config: config.Config{
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

			config: config.Config{
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

			t.Setenv("APP_PORT", testCase.config.AppPort)
			t.Setenv("SERVICE_NAME", testCase.config.ServiceName)
			if testCase.config.InstanceID != "" {
				t.Setenv(
					"INSTANCE_ID",
					testCase.config.InstanceID,
				)
			} else {
				_ = os.Unsetenv("INSTANCE_ID")
			}

			cfg, err := config.LoadConfig()

			require.NoError(t, err)

			healthCheckService := service.NewHealthCheckService(
				cfg.ServiceName,
				cfg.InstanceID,
			)

			healthCheckHandler := handler.NewHealthCheckHandler(
				healthCheckService,
			)

			router := api.NewRouter(
				cfg.AppPort,
				healthCheckHandler,
			)

			req := httptest.NewRequest(
				http.MethodGet,
				"/health-check",
				nil,
			)

			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, req)

			require.Equal(
				t,
				testCase.expectedHTTPStatusCode,
				recorder.Code,
			)

			var response model.HealthCheckResponse

			err = json.Unmarshal(
				recorder.Body.Bytes(),
				&response,
			)

			require.NoError(t, err)

			assert.Equal(
				t,
				testCase.expectedMessage,
				response.Message,
			)

			assert.Equal(
				t,
				testCase.expectedServiceName,
				response.ServiceName,
			)

			if testCase.expectGeneratedUUID {
				assert.NotEmpty(
					t,
					response.InstanceID,
				)
			} else {
				assert.Equal(
					t,
					testCase.expectedInstanceID,
					response.InstanceID,
				)
			}
		})
	}
}
