package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/huypham67/bookmark-management/internal/api"
	"github.com/huypham67/bookmark-management/internal/config"
	"github.com/huypham67/bookmark-management/internal/handler"
	"github.com/huypham67/bookmark-management/internal/repository"
	"github.com/huypham67/bookmark-management/internal/service"
	"github.com/huypham67/bookmark-management/internal/utils"
	"github.com/huypham67/bookmark-management/pkg/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedirectToURLEndpoint(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                   string
		appConfig              config.Config
		code                   string
		originalURL            string
		setupRedis             func(*redis.MockRedis, string, string) error
		expectedHTTPStatusCode int
		expectRedirect         bool
	}{
		{
			name: "should redirect successfully when short code exists",
			appConfig: config.Config{
				AppPort:     "8080",
				ServiceName: "bookmark-service",
				InstanceID:  "instance-1",
			},
			code:        "abc123d",
			originalURL: "https://example.com/very/long/path",
			setupRedis: func(mr *redis.MockRedis, code, url string) error {
				return mr.Server.Set(code, url)
			},
			expectedHTTPStatusCode: http.StatusFound,
			expectRedirect:         true,
		},
		{
			name: "should return 404 when short code does not exist",
			appConfig: config.Config{
				AppPort:     "8081",
				ServiceName: "api-service",
				InstanceID:  "api-1",
			},
			code:        "notfound",
			originalURL: "",
			setupRedis: func(mr *redis.MockRedis, code, url string) error {
				// Don't set anything in Redis
				return nil
			},
			expectedHTTPStatusCode: http.StatusNotFound,
			expectRedirect:         false,
		},
		{
			name: "should handle URL with query parameters",
			appConfig: config.Config{
				AppPort:     "8082",
				ServiceName: "test-service",
				InstanceID:  "test-1",
			},
			code:        "query01",
			originalURL: "https://example.com?key=value&foo=bar",
			setupRedis: func(mr *redis.MockRedis, code, url string) error {
				return mr.Server.Set(code, url)
			},
			expectedHTTPStatusCode: http.StatusFound,
			expectRedirect:         true,
		},
		{
			name: "should handle URL with hash fragment",
			appConfig: config.Config{
				AppPort:     "8083",
				ServiceName: "hash-service",
				InstanceID:  "hash-1",
			},
			code:        "hash001",
			originalURL: "https://example.com#section",
			setupRedis: func(mr *redis.MockRedis, code, url string) error {
				return mr.Server.Set(code, url)
			},
			expectedHTTPStatusCode: http.StatusFound,
			expectRedirect:         true,
		},
		{
			name: "should handle very long URL",
			appConfig: config.Config{
				AppPort:     "8084",
				ServiceName: "long-service",
				InstanceID:  "long-1",
			},
			code:        "verylong",
			originalURL: "https://example.com/very/long/path/with/many/segments?param1=value1&param2=value2&param3=value3",
			setupRedis: func(mr *redis.MockRedis, code, url string) error {
				return mr.Server.Set(code, url)
			},
			expectedHTTPStatusCode: http.StatusFound,
			expectRedirect:         true,
		},
		{
			name: "should handle different service names",
			appConfig: config.Config{
				AppPort:     "8085",
				ServiceName: "different-service",
				InstanceID:  "diff-1",
			},
			code:        "diff01",
			originalURL: "https://github.com/huypham67/bookmark-management",
			setupRedis: func(mr *redis.MockRedis, code, url string) error {
				return mr.Server.Set(code, url)
			},
			expectedHTTPStatusCode: http.StatusFound,
			expectRedirect:         true,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// 1. Setup Mock Redis
			mockRedis := redis.NewMockRedis(t)

			// 2. Setup Redis with data if needed
			err := tc.setupRedis(mockRedis, tc.code, tc.originalURL)
			require.NoError(t, err)

			// 3. Create Link Repository
			linkRepo := repository.NewLinkRepository(&redis.RedisClient{Client: mockRedis.Client})

			// 4. Create Link Service
			linkService := service.NewLinkService(linkRepo, utils.NewCodeGenerator())

			// 5. Create Link Handler
			linkHandler := handler.NewLinkHandler(linkService)

			// 6. Initialize Router
			router := api.NewRouter(tc.appConfig.AppPort)
			api.RegisterLinkRoutes(router.GroupV1(), linkHandler)

			// 7. Execute Request
			req := httptest.NewRequest(
				http.MethodGet,
				"/api/v1/links/redirect/"+tc.code,
				nil,
			)

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			// 8. Assertions
			assert.Equal(t, tc.expectedHTTPStatusCode, recorder.Code)

			if tc.expectRedirect {
				// Verify redirect location
				assert.Equal(t, tc.originalURL, recorder.Header().Get("Location"))
			} else {
				// Verify error response
				var errRes map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &errRes)
				require.NoError(t, err)
				assert.Equal(t, "Short link not found", errRes["error"])
			}
		})
	}
}
