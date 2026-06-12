package link

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/huypham67/bookmark-common/pkg/dbutils"
	"github.com/huypham67/bookmark-service/internal/service/link/mocks"
	"github.com/stretchr/testify/assert"
)

func TestHandler_RedirectToURL(t *testing.T) {
	t.Parallel()

	type expected struct {
		statusCode int
		location   string
		body       string
	}

	testCases := []struct {
		name      string
		code      string
		setupMock func(context.Context, *mocks.Service)
		expected  expected
	}{
		{
			name: "should redirect to original URL successfully",
			code: "abc1234",
			setupMock: func(ctx context.Context, mockService *mocks.Service) {
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
			setupMock: func(ctx context.Context, mockService *mocks.Service) {
				mockService.
					On(
						"GetOriginalURL",
						ctx,
						"missing",
					).
					Return("", dbutils.ErrRecordNotFoundType).
					Once()
			},
			expected: expected{
				statusCode: http.StatusNotFound,
				body:       "Short link not found",
			},
		},
		{
			name: "should return 400 when code parameter is missing",
			code: "",
			setupMock: func(ctx context.Context, mockService *mocks.Service) {
				// No mock setup needed for invalid request
			},
			expected: expected{
				statusCode: http.StatusBadRequest,
				body:       "Invalid request",
			},
		},
		{
			name: "should return 500 when service returns unexpected error",
			code: "abc1234",
			setupMock: func(ctx context.Context, mockService *mocks.Service) {
				mockService.
					On(
						"GetOriginalURL",
						ctx,
						"abc1234",
					).
					Return("", assert.AnError).
					Once()
			},
			expected: expected{
				statusCode: http.StatusInternalServerError,
				body:       "Internal Server Error",
			},
		},
	}
	for _, tc := range testCases {

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gin.SetMode(gin.TestMode)
			mockSvc := mocks.NewService(t)

			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			httpRequest := httptest.NewRequest(http.MethodGet, "/links/"+tc.code, nil)
			ctx.Request = httpRequest

			// Only set params if code is not empty
			if tc.code != "" {
				ctx.Params = []gin.Param{{Key: "code", Value: tc.code}}
			}

			tc.setupMock(ctx, mockSvc)

			handler := NewHandler(mockSvc)

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
