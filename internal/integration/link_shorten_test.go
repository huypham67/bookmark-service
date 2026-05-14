package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/huypham67/bookmark-management/internal/api"
	"github.com/huypham67/bookmark-management/internal/config"
	"github.com/huypham67/bookmark-management/internal/dto/request"
	"github.com/huypham67/bookmark-management/internal/dto/response"
	"github.com/huypham67/bookmark-management/internal/handler"
	"github.com/huypham67/bookmark-management/internal/repository"
	"github.com/huypham67/bookmark-management/internal/service"
	"github.com/huypham67/bookmark-management/internal/utils"
	"github.com/huypham67/bookmark-management/pkg/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShortenURLEndpoint(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                   string
		appConfig              config.Config
		requestURL             string
		requestExp             int64
		expectedHTTPStatusCode int
		expectedMessage        string
		expectError            bool
		validateCodeFormat     bool
	}{
		{
			name: "should successfully shorten URL with valid request",
			appConfig: config.Config{
				AppPort:     "8080",
				ServiceName: "bookmark-service",
				InstanceID:  "instance-1",
			},
			requestURL:             "https://github.com/user/repository/issues?state=open&labels=bug",
			requestExp:             3600,
			expectedHTTPStatusCode: http.StatusOK,
			expectedMessage:        "Shorten URL generated successfully",
			expectError:            false,
			validateCodeFormat:     true,
		},
		{
			name: "should handle short expiration time",
			appConfig: config.Config{
				AppPort:     "8080",
				ServiceName: "bookmark-service",
				InstanceID:  "instance-2",
			},
			requestURL:             "https://example.com/short-lived",
			requestExp:             60,
			expectedHTTPStatusCode: http.StatusOK,
			expectedMessage:        "Shorten URL generated successfully",
			expectError:            false,
			validateCodeFormat:     true,
		},
		{
			name: "should handle long expiration time",
			appConfig: config.Config{
				AppPort:     "8081",
				ServiceName: "test-service",
				InstanceID:  "test-instance",
			},
			requestURL:             "https://example.com/long-lived",
			requestExp:             86400,
			expectedHTTPStatusCode: http.StatusOK,
			expectedMessage:        "Shorten URL generated successfully",
			expectError:            false,
			validateCodeFormat:     true,
		},
		{
			name: "should handle URL with query parameters",
			appConfig: config.Config{
				AppPort:     "8082",
				ServiceName: "api-service",
				InstanceID:  "api-1",
			},
			requestURL:             "https://example.com?key=value&foo=bar",
			requestExp:             3600,
			expectedHTTPStatusCode: http.StatusOK,
			expectedMessage:        "Shorten URL generated successfully",
			expectError:            false,
			validateCodeFormat:     true,
		},
		{
			name: "should handle URL with zero expiration",
			appConfig: config.Config{
				AppPort:     "8083",
				ServiceName: "service-zero",
				InstanceID:  "zero-1",
			},
			requestURL:             "https://example.com/no-exp",
			requestExp:             0,
			expectedHTTPStatusCode: http.StatusOK,
			expectedMessage:        "Shorten URL generated successfully",
			expectError:            false,
			validateCodeFormat:     true,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// 1. Setup Mock Redis
			mockRedis := redis.NewMockRedis(t)

			// 2. Create Link Repository
			linkRepo := repository.NewLinkRepository(&redis.RedisClient{Client: mockRedis.Client})

			// 3. Create Code Generator
			codeGenerator := utils.NewCodeGenerator()

			// 4. Create Link Service
			linkService := service.NewLinkService(linkRepo, codeGenerator)

			// 5. Create Link Handler
			linkHandler := handler.NewLinkHandler(linkService)

			// 6. Initialize Router
			router := api.NewRouter(tc.appConfig.AppPort)
			api.RegisterLinkRoutes(router.GroupV1(), linkHandler)

			// 7. Build Request
			reqBody := request.ShortenURLRequest{
				Url: tc.requestURL,
				Exp: tc.requestExp,
			}
			bodyBytes, err := json.Marshal(reqBody)
			require.NoError(t, err)

			// 8. Execute Request
			req := httptest.NewRequest(
				http.MethodPost,
				"/api/v1/links/shorten",
				bytes.NewReader(bodyBytes),
			)
			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			// 9. Assertions
			assert.Equal(t, tc.expectedHTTPStatusCode, recorder.Code)

			if tc.expectError {
				var errRes map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &errRes)
				require.NoError(t, err)
				assert.NotEmpty(t, errRes["error"])
			} else {
				var res response.ShortenURLResponse
				err := json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)

				assert.Equal(t, tc.expectedMessage, res.Message)
				assert.NotEmpty(t, res.Code)

				if tc.validateCodeFormat {
					// Verify code is exactly 7 characters and alphanumeric
					assert.Equal(t, 7, len(res.Code))
					for _, ch := range res.Code {
						assert.True(t,
							(ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9'),
							"code should only contain alphanumeric characters")
					}
				}
			}
		})
	}
}
