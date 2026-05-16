package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/huypham67/bookmark-service/internal/dto/request"
	"github.com/huypham67/bookmark-service/internal/service/mocks"
	"github.com/stretchr/testify/assert"
)

func TestLinkHandler_ShortenURL(t *testing.T) {
	t.Parallel()

	type expected struct {
		statusCode int
		body       string
	}

	testCases := []struct {
		name        string
		requestBody string
		setupMock   func(context.Context, *mocks.LinkService)
		expected    expected
	}{
		{
			name: "should return 200 when shorten URL succeeds",
			requestBody: `{
				"url": "https://www.google.com",
				"exp": 3600
			}`,
			setupMock: func(ctx context.Context, mockService *mocks.LinkService) {
				mockService.
					On(
						"ShortenURL",
						ctx,
						request.ShortenURLRequest{
							Url: "https://www.google.com",
							Exp: 3600,
						},
					).
					Return("abc1234", nil).
					Once()
			},
			expected: expected{
				statusCode: http.StatusOK,
				body:       "abc1234",
			},
		}, {
			name: "should return 400 when validation fails",
			requestBody: `{
				"url": "invalid-url",
				"exp": 3600
			}`,
			setupMock: func(ctx context.Context, mockService *mocks.LinkService) {
				// No need to set up mock for validation failure
			},
			expected: expected{
				statusCode: http.StatusBadRequest,
				body:       "Invalid request body",
			},
		}, {
			name: "should return 500 when service returns error",
			requestBody: `{
				"url": "https://www.google.com",
				"exp": 3600
			}`,
			setupMock: func(ctx context.Context, mockService *mocks.LinkService) {
				mockService.
					On(
						"ShortenURL",
						ctx,
						request.ShortenURLRequest{
							Url: "https://www.google.com",
							Exp: 3600,
						},
					).
					Return("", errors.New("redis error")).
					Once()
			},
			expected: expected{
				statusCode: http.StatusInternalServerError,
				body:       "Internal Server Error",
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gin.SetMode(gin.TestMode)
			mockSvc := mocks.NewLinkService(t)

			recorder := httptest.NewRecorder()

			ctx, _ := gin.CreateTestContext(recorder)

			httpRequest := httptest.NewRequest(
				http.MethodPost,
				"/links/shorten",
				strings.NewReader(tc.requestBody),
			)

			httpRequest.Header.Set("Content-Type", "application/json")
			ctx.Request = httpRequest

			tc.setupMock(ctx, mockSvc)

			handler := NewLinkHandler(mockSvc)

			handler.ShortenURL(ctx)

			assert.Equal(t, tc.expected.statusCode, recorder.Code)
			assert.Equal(t, "application/json; charset=utf-8", recorder.Header().Get("Content-Type"))
			assert.Contains(t, recorder.Body.String(), tc.expected.body)

			mockSvc.AssertExpectations(t)
		})
	}

}

func TestLinkHandler_RedirectToURL(t *testing.T) {
	t.Parallel()

	type expected struct {
		statusCode int
		location   string
		body       string
	}

	testCases := []struct {
		name      string
		code      string
		setupMock func(context.Context, *mocks.LinkService)
		expected  expected
	}{
		{
			name: "should redirect to original URL successfully",
			code: "abc1234",
			setupMock: func(ctx context.Context, mockService *mocks.LinkService) {
				mockService.
					On(
						"GetOriginalURL",
						ctx,
						"abc1234",
					).
					Return("https://www.google.com", nil).
					Once()
			},
			expected: expected{
				statusCode: http.StatusFound,
				location:   "https://www.google.com",
			},
		},
		{
			name: "should return 404 when shorten code does not exist",
			code: "missing",
			setupMock: func(ctx context.Context, mockService *mocks.LinkService) {
				mockService.
					On(
						"GetOriginalURL",
						ctx,
						"missing",
					).
					Return("", errors.New("not found")).
					Once()
			},
			expected: expected{
				statusCode: http.StatusNotFound,
				body:       "Short link not found",
			},
		},
	}
	for _, tc := range testCases {

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gin.SetMode(gin.TestMode)
			mockSvc := mocks.NewLinkService(t)

			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			httpRequest := httptest.NewRequest(http.MethodGet, "/links/"+tc.code, nil)
			ctx.Request = httpRequest
			ctx.Params = []gin.Param{{Key: "code", Value: tc.code}}

			tc.setupMock(ctx, mockSvc)

			handler := NewLinkHandler(mockSvc)

			handler.RedirectToURL(ctx)

			assert.Equal(t, tc.expected.statusCode, recorder.Code)
			if tc.expected.location != "" {
				assert.Equal(t, tc.expected.location, recorder.Header().Get("Location"))
				assert.Equal(t, "text/html; charset=utf-8", recorder.Header().Get("Content-Type"))
			}
			if tc.expected.body != "" {
				assert.Contains(t, recorder.Body.String(), tc.expected.body)
			}

			mockSvc.AssertExpectations(t)
		})
	}
}
