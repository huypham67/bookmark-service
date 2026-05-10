package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/huypham67/bookmark-management/infrastructure/redis"
	"github.com/huypham67/bookmark-management/internal/api"
	"github.com/huypham67/bookmark-management/internal/config"
	"github.com/huypham67/bookmark-management/internal/dto/request"
	"github.com/huypham67/bookmark-management/internal/dto/response"
	healthHandler "github.com/huypham67/bookmark-management/internal/handler/health"
	linkHandler "github.com/huypham67/bookmark-management/internal/handler/link"
	linkRepository "github.com/huypham67/bookmark-management/internal/repository"
	healthService "github.com/huypham67/bookmark-management/internal/service/health"
	linkService "github.com/huypham67/bookmark-management/internal/service/link"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShortenURLEndpoint(t *testing.T) {
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

	cfg := &config.AppConfig{
		AppPort:     "8080",
		ServiceName: "bookmark-service",
		InstanceID:  "instance-1",
	}

	testCases := []struct {
		name                   string
		buildRequest           func() (*request.ShortenURLRequest, error)
		verifyResponse         func(*response.ShortenURLResponse, *httptest.ResponseRecorder) bool
		expectedHTTPStatusCode int
		shouldSucceed          bool
		setupRedis             func(*miniredis.Miniredis) error
		testType               string // "basic", "structure"
		validateResponseFields func(t *testing.T, res *response.ShortenURLResponse)
	}{
		// Basic functionality tests
		{
			name: "should successfully shorten URL with valid request",
			buildRequest: func() (*request.ShortenURLRequest, error) {
				return &request.ShortenURLRequest{
					Url: "https://github.com/user/repository/issues?state=open&labels=bug",
					Exp: 3600,
				}, nil
			},
			verifyResponse: func(res *response.ShortenURLResponse, rec *httptest.ResponseRecorder) bool {
				return assert.NotEmpty(t, res.Code) &&
					assert.Len(t, res.Code, 7) &&
					assert.Equal(t, "Shorten URL generated successfully", res.Message) &&
					assert.Equal(t, http.StatusOK, rec.Code)
			},
			expectedHTTPStatusCode: http.StatusOK,
			shouldSucceed:          true,
			setupRedis: func(mr *miniredis.Miniredis) error {
				return nil
			},
			testType: "basic",
		},
		{
			name: "should handle short expiration time",
			buildRequest: func() (*request.ShortenURLRequest, error) {
				return &request.ShortenURLRequest{
					Url: "https://example.com/short-lived",
					Exp: 60,
				}, nil
			},
			verifyResponse: func(res *response.ShortenURLResponse, rec *httptest.ResponseRecorder) bool {
				return assert.NotEmpty(t, res.Code) &&
					assert.Equal(t, "Shorten URL generated successfully", res.Message)
			},
			expectedHTTPStatusCode: http.StatusOK,
			shouldSucceed:          true,
			setupRedis: func(mr *miniredis.Miniredis) error {
				return nil
			},
			testType: "basic",
		},
		{
			name: "should handle long expiration time",
			buildRequest: func() (*request.ShortenURLRequest, error) {
				return &request.ShortenURLRequest{
					Url: "https://example.com/long-lived",
					Exp: 86400 * 365,
				}, nil
			},
			verifyResponse: func(res *response.ShortenURLResponse, rec *httptest.ResponseRecorder) bool {
				return assert.NotEmpty(t, res.Code) &&
					assert.Equal(t, "Shorten URL generated successfully", res.Message)
			},
			expectedHTTPStatusCode: http.StatusOK,
			shouldSucceed:          true,
			setupRedis: func(mr *miniredis.Miniredis) error {
				return nil
			},
			testType: "basic",
		},
		{
			name: "should handle URL with query parameters",
			buildRequest: func() (*request.ShortenURLRequest, error) {
				return &request.ShortenURLRequest{
					Url: "https://example.com/search?q=golang&sort=stars&order=desc",
					Exp: 7200,
				}, nil
			},
			verifyResponse: func(res *response.ShortenURLResponse, rec *httptest.ResponseRecorder) bool {
				return assert.NotEmpty(t, res.Code) &&
					assert.Equal(t, "Shorten URL generated successfully", res.Message)
			},
			expectedHTTPStatusCode: http.StatusOK,
			shouldSucceed:          true,
			setupRedis: func(mr *miniredis.Miniredis) error {
				return nil
			},
			testType: "basic",
		},
		{
			name: "should handle URL with hash fragment",
			buildRequest: func() (*request.ShortenURLRequest, error) {
				return &request.ShortenURLRequest{
					Url: "https://example.com/docs#section-installation",
					Exp: 3600,
				}, nil
			},
			verifyResponse: func(res *response.ShortenURLResponse, rec *httptest.ResponseRecorder) bool {
				return assert.NotEmpty(t, res.Code) &&
					assert.Equal(t, "Shorten URL generated successfully", res.Message)
			},
			expectedHTTPStatusCode: http.StatusOK,
			shouldSucceed:          true,
			setupRedis: func(mr *miniredis.Miniredis) error {
				return nil
			},
			testType: "basic",
		},
		{
			name: "should handle very long URL",
			buildRequest: func() (*request.ShortenURLRequest, error) {
				return &request.ShortenURLRequest{
					Url: "https://example.com/api/v1/resources/items/12345/details?nested=true&format=json&expand=all&include=metadata,history,relationships&filter=active&page=1&limit=100&sort=-created_at&foo=bar&baz=qux",
					Exp: 3600,
				}, nil
			},
			verifyResponse: func(res *response.ShortenURLResponse, rec *httptest.ResponseRecorder) bool {
				return assert.NotEmpty(t, res.Code) &&
					assert.Equal(t, "Shorten URL generated successfully", res.Message)
			},
			expectedHTTPStatusCode: http.StatusOK,
			shouldSucceed:          true,
			setupRedis: func(mr *miniredis.Miniredis) error {
				return nil
			},
			testType: "basic",
		},
		{
			name: "should handle empty URL by still generating a short code",
			buildRequest: func() (*request.ShortenURLRequest, error) {
				return &request.ShortenURLRequest{
					Url: "",
					Exp: 3600,
				}, nil
			},
			verifyResponse: func(res *response.ShortenURLResponse, rec *httptest.ResponseRecorder) bool {
				return assert.NotEmpty(t, res.Code) &&
					assert.Equal(t, "Shorten URL generated successfully", res.Message)
			},
			expectedHTTPStatusCode: http.StatusOK,
			shouldSucceed:          true,
			setupRedis: func(mr *miniredis.Miniredis) error {
				return nil
			},
			testType: "basic",
		},
		{
			name: "should handle multiple URLs without collision",
			buildRequest: func() (*request.ShortenURLRequest, error) {
				return &request.ShortenURLRequest{
					Url: "https://different-url-for-collision-test.com/path123",
					Exp: 3600,
				}, nil
			},
			verifyResponse: func(res *response.ShortenURLResponse, rec *httptest.ResponseRecorder) bool {
				return assert.NotEmpty(t, res.Code) &&
					assert.Equal(t, "Shorten URL generated successfully", res.Message)
			},
			expectedHTTPStatusCode: http.StatusOK,
			shouldSucceed:          true,
			setupRedis: func(mr *miniredis.Miniredis) error {
				return nil
			},
			testType: "basic",
		},
		// Response structure tests
		{
			name: "response should contain both code and message fields",
			buildRequest: func() (*request.ShortenURLRequest, error) {
				return &request.ShortenURLRequest{
					Url: "https://example.com/test",
					Exp: 3600,
				}, nil
			},
			expectedHTTPStatusCode: http.StatusOK,
			shouldSucceed:          true,
			setupRedis: func(mr *miniredis.Miniredis) error {
				return nil
			},
			testType: "structure",
			validateResponseFields: func(t *testing.T, res *response.ShortenURLResponse) {
				assert.NotEmpty(t, res.Code)
				assert.NotEmpty(t, res.Message)
				assert.Equal(t, 7, len(res.Code))
			},
		},
		{
			name: "code field should be exactly 7 characters",
			buildRequest: func() (*request.ShortenURLRequest, error) {
				return &request.ShortenURLRequest{
					Url: "https://another-test-url.com/path",
					Exp: 7200,
				}, nil
			},
			expectedHTTPStatusCode: http.StatusOK,
			shouldSucceed:          true,
			setupRedis: func(mr *miniredis.Miniredis) error {
				return nil
			},
			testType: "structure",
			validateResponseFields: func(t *testing.T, res *response.ShortenURLResponse) {
				codeLen := len(res.Code)
				assert.Equal(t, 7, codeLen)
				for _, char := range res.Code {
					assert.True(t, (char >= 'a' && char <= 'z') ||
						(char >= 'A' && char <= 'Z') ||
						(char >= '0' && char <= '9'),
						"code should only contain alphanumeric characters")
				}
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			// Setup Redis state if needed
			err := testCase.setupRedis(mr)
			require.NoError(t, err)

			// Create link service with real repository
			linkRepo := linkRepository.NewLinkRepository(redisClient)
			linkSvc := linkService.NewLinkService(linkRepo)
			linkHandlerInstance := linkHandler.NewLinkHandler(linkSvc)

			// Create health check service
			healthCheckSvc := healthService.NewHealthCheckService(
				cfg.ServiceName,
				cfg.InstanceID,
				redisClient,
			)
			healthCheckHandlerInstance := healthHandler.NewHealthCheckHandler(
				healthCheckSvc,
			)

			router := api.NewRouter(
				cfg.AppPort,
				healthCheckHandlerInstance,
				linkHandlerInstance,
			)

			// Build request
			req, err := testCase.buildRequest()
			require.NoError(t, err)

			// Marshal request body
			body, err := json.Marshal(req)
			require.NoError(t, err)

			// Create HTTP request
			httpReq := httptest.NewRequest(
				http.MethodPost,
				"/api/v1/links/shorten-url",
				bytes.NewReader(body),
			)
			httpReq.Header.Set("Content-Type", "application/json")

			// Execute request
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httpReq)

			// Verify response
			require.Equal(t, testCase.expectedHTTPStatusCode, recorder.Code)

			if testCase.shouldSucceed {
				var res response.ShortenURLResponse
				err := json.Unmarshal(recorder.Body.Bytes(), &res)
				require.NoError(t, err)

				// Run custom verification based on test type
				if testCase.testType == "basic" {
					_ = testCase.verifyResponse(&res, recorder)
					// Verify the code was actually saved in Redis
					storedURL, err := redisClient.Get(res.Code)
					require.NoError(t, err)
					assert.Equal(t, req.Url, storedURL)
				} else if testCase.testType == "structure" {
					testCase.validateResponseFields(t, &res)
				}
			} else {
				// For failed cases, verify error response
				var errResponse map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &errResponse)
				require.NoError(t, err)
				assert.NotNil(t, errResponse["error"])
			}

			// Clear Redis for next test
			mr.FlushDB()
		})
	}
}
