package bookmark

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/huypham67/bookmark-common/pkg/jwt"
	"github.com/huypham67/bookmark-service/internal/service/bookmark"
	"github.com/huypham67/bookmark-service/internal/service/bookmark/mocks"
	"github.com/stretchr/testify/assert"
)

func TestHandler_Delete(t *testing.T) {
	t.Parallel()

	type expected struct {
		statusCode   int
		bodyContains string
	}

	testCases := []struct {
		name        string
		bookmarkID  string
		setupClaims func(*gin.Context)
		setupMock   func(context.Context, *mocks.Service)
		expected    expected
	}{
		{
			name:       "should return 200 when bookmark is deleted successfully",
			bookmarkID: "bm-123",
			setupClaims: func(ctx *gin.Context) {
				ctx.Set("claims", &jwt.CustomClaims{
					UserID: "user-id-123",
				})
			},
			setupMock: func(ctx context.Context, mockSvc *mocks.Service) {
				mockSvc.
					On(
						"Delete",
						ctx,
						"user-id-123",
						"bm-123",
					).
					Return(nil).
					Once()
			},
			expected: expected{
				statusCode:   http.StatusOK,
				bodyContains: "Success",
			},
		},
		{
			name:       "should return 401 when user ID is not found in context",
			bookmarkID: "bm-123",
			setupClaims: func(ctx *gin.Context) {
				// No claims set
			},
			setupMock: func(ctx context.Context, mockSvc *mocks.Service) {
				// No mock setup needed
			},
			expected: expected{
				statusCode:   http.StatusUnauthorized,
				bodyContains: "Unauthorized",
			},
		},
		{
			name:       "should return 404 when bookmark is not found",
			bookmarkID: "nonexistent-bm",
			setupClaims: func(ctx *gin.Context) {
				ctx.Set("claims", &jwt.CustomClaims{
					UserID: "user-id-123",
				})
			},
			setupMock: func(ctx context.Context, mockSvc *mocks.Service) {
				mockSvc.
					On(
						"Delete",
						ctx,
						"user-id-123",
						"nonexistent-bm",
					).
					Return(bookmark.ErrBookmarkNotFound).
					Once()
			},
			expected: expected{
				statusCode:   http.StatusNotFound,
				bodyContains: "Bookmark not found",
			},
		},
		{
			name:       "should return 400 when service returns bad request error",
			bookmarkID: "bm-123",
			setupClaims: func(ctx *gin.Context) {
				ctx.Set("claims", &jwt.CustomClaims{
					UserID: "user-id-123",
				})
			},
			setupMock: func(ctx context.Context, mockSvc *mocks.Service) {
				mockSvc.
					On(
						"Delete",
						ctx,
						"user-id-123",
						"bm-123",
					).
					Return(bookmark.ErrBadRequest).
					Once()
			},
			expected: expected{
				statusCode:   http.StatusBadRequest,
				bodyContains: "invalid request",
			},
		},
		{
			name:       "should return 500 when service returns unexpected error",
			bookmarkID: "bm-123",
			setupClaims: func(ctx *gin.Context) {
				ctx.Set("claims", &jwt.CustomClaims{
					UserID: "user-id-123",
				})
			},
			setupMock: func(ctx context.Context, mockSvc *mocks.Service) {
				mockSvc.
					On(
						"Delete",
						ctx,
						"user-id-123",
						"bm-123",
					).
					Return(errors.New("database error")).
					Once()
			},
			expected: expected{
				statusCode:   http.StatusInternalServerError,
				bodyContains: "Internal Server Error",
			},
		},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gin.SetMode(gin.TestMode)
			mockSvc := new(mocks.Service)

			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)

			httpRequest := httptest.NewRequest(
				http.MethodDelete,
				"/v1/bookmarks/"+tc.bookmarkID,
				nil,
			)

			ctx.Request = httpRequest
			ctx.Params = []gin.Param{{Key: "id", Value: tc.bookmarkID}}

			tc.setupClaims(ctx)
			tc.setupMock(ctx, mockSvc)

			handler := NewHandler(mockSvc)
			handler.Delete(ctx)

			assert.Equal(t, tc.expected.statusCode, recorder.Code)
			assert.Equal(t, "application/json; charset=utf-8", recorder.Header().Get("Content-Type"))
			assert.Contains(t, recorder.Body.String(), tc.expected.bodyContains)

			mockSvc.AssertExpectations(t)
		})
	}
}
