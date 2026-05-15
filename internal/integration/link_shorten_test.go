package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/huypham67/bookmark-management/internal/api"
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

	type expected struct {
		statusCode   int
		bodyContains string
	}

	testCases := []struct {
		name        string
		requestBody string
		setupRedis  func(*redis.MockRedis)
		expected    expected
	}{
		{
			name: "should return 200 when request is valid",
			requestBody: `
			{
				"url": "https://example.com",
				"exp": 3600
			}
			`,
			setupRedis: func(
				mockRedis *redis.MockRedis,
			) {
			},
			expected: expected{
				statusCode:   http.StatusOK,
				bodyContains: "Shorten URL generated successfully",
			},
		},
		{
			name:        "should return 400 when request body is invalid JSON",
			requestBody: `{invalid json}`,
			setupRedis: func(
				mockRedis *redis.MockRedis,
			) {
			},
			expected: expected{
				statusCode:   http.StatusBadRequest,
				bodyContains: "Invalid request body",
			},
		},
		{
			name: "should return 400 when validation fails",
			requestBody: `
			{
				"url": "",
				"exp": 3600
			}
			`,
			setupRedis: func(
				mockRedis *redis.MockRedis,
			) {
			},
			expected: expected{
				statusCode:   http.StatusBadRequest,
				bodyContains: "Validation failed",
			},
		},
		{
			name: "should return 500 when redis connection fails",
			requestBody: `
			{
				"url": "https://example.com",
				"exp": 3600
			}
			`,
			setupRedis: func(
				mockRedis *redis.MockRedis,
			) {
				mockRedis.Close()
			},
			expected: expected{
				statusCode:   http.StatusInternalServerError,
				bodyContains: "Internal Server Error",
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockRedis := redis.NewMockRedis(t)
			tc.setupRedis(mockRedis)
			linkRepository := repository.NewLinkRepository(&redis.RedisClient{Client: mockRedis.Client})
			codeGenerator := utils.NewCodeGenerator()

			linkService := service.NewLinkService(linkRepository, codeGenerator)

			linkHandler := handler.NewLinkHandler(linkService)

			router := api.NewRouter()
			api.RegisterLinkRoutes(router.GroupV1(), linkHandler)

			httpRequest := httptest.NewRequest(http.MethodPost, "/api/v1/links/shorten", bytes.NewBufferString(tc.requestBody))

			httpRequest.Header.Set("Content-Type", "application/json")
			httpRecorder := httptest.NewRecorder()
			router.ServeHTTP(httpRecorder, httpRequest)

			assert.Equal(t, tc.expected.statusCode, httpRecorder.Code)
			assert.Equal(t, "application/json; charset=utf-8", httpRecorder.Header().Get("Content-Type"))
			assert.Contains(t, httpRecorder.Body.String(), tc.expected.bodyContains)

			if tc.expected.statusCode == http.StatusOK {
				var actual response.ShortenURLResponse

				err := json.Unmarshal(httpRecorder.Body.Bytes(), &actual)

				require.NoError(t, err)
				assert.NotEmpty(t, actual.Code)
			}
		})
	}
}
