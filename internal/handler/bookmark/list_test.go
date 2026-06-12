package bookmark

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/huypham67/bookmark-common/pkg/jwt"
	bookmarkDTO "github.com/huypham67/bookmark-service/internal/dto/bookmark"
	"github.com/huypham67/bookmark-service/internal/model"
	"github.com/huypham67/bookmark-service/internal/service/bookmark"
	"github.com/huypham67/bookmark-service/internal/service/bookmark/mocks"
	"github.com/stretchr/testify/assert"
)

func TestHandler_List(t *testing.T) {
	t.Parallel()

	type expected struct {
		statusCode   int
		bodyContains string
	}

	testCases := []struct {
		name        string
		queryParams string
		setupClaims func(*gin.Context)
		setupMock   func(context.Context, *mocks.Service)
		expected    expected
	}{
		{
			name:        "should return 200 with bookmarks list",
			queryParams: "?page=1&limit=10&sort=created_at",
			setupClaims: func(ctx *gin.Context) {
				ctx.Set("claims", &jwt.CustomClaims{
					UserID: "user-id-123",
				})
			},
			setupMock: func(ctx context.Context, mockSvc *mocks.Service) {
				bookmarks := []*model.Bookmark{
					{
						BaseModel: model.BaseModel{
							ID: "bm-1",
						},
						Description: "Bookmark 1",
						URL:         "https://example1.com",
						Code:        "abc123",
						UserID:      "user-id-123",
					},
					{
						BaseModel: model.BaseModel{
							ID: "bm-2",
						},
						Description: "Bookmark 2",
						URL:         "https://example2.com",
						Code:        "def456",
						UserID:      "user-id-123",
					},
				}

				mockSvc.
					On(
						"List",
						ctx,
						"user-id-123",
						&bookmarkDTO.ListBookmarksRequest{
							Page:  1,
							Limit: 10,
							Sort:  "created_at",
						},
					).
					Return(
						bookmarks,
						&bookmark.PaginationResult{
							Page:  1,
							Limit: 10,
							Total: 2,
						},
						nil,
					).
					Once()
			},
			expected: expected{
				statusCode:   http.StatusOK,
				bodyContains: "Bookmarks retrieved successfully!",
			},
		},
		{
			name:        "should return 200 with empty bookmarks list",
			queryParams: "?page=1&limit=10",
			setupClaims: func(ctx *gin.Context) {
				ctx.Set("claims", &jwt.CustomClaims{
					UserID: "user-id-456",
				})
			},
			setupMock: func(ctx context.Context, mockSvc *mocks.Service) {
				emptyBookmarks := make([]*model.Bookmark, 0)

				mockSvc.
					On(
						"List",
						ctx,
						"user-id-456",
						&bookmarkDTO.ListBookmarksRequest{
							Page:  1,
							Limit: 10,
							Sort:  "created_at",
						},
					).
					Return(
						emptyBookmarks,
						&bookmark.PaginationResult{
							Page:  1,
							Limit: 10,
							Total: 0,
						},
						nil,
					).
					Once()
			},
			expected: expected{
				statusCode:   http.StatusOK,
				bodyContains: "Bookmarks retrieved successfully!",
			},
		},
		{
			name:        "should return 401 when user ID is not found in context",
			queryParams: "?page=1&limit=10",
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
			name:        "should return 400 when query parameters are invalid",
			queryParams: "?page=-1&limit=200&sort=invalid",
			setupClaims: func(ctx *gin.Context) {
				ctx.Set("claims", &jwt.CustomClaims{
					UserID: "user-id-123",
				})
			},
			setupMock: func(ctx context.Context, mockSvc *mocks.Service) {
				// No mock setup needed for validation failure
			},
			expected: expected{
				statusCode:   http.StatusBadRequest,
				bodyContains: "Invalid query parameters",
			},
		},
		{
			name:        "should return 500 when service returns error",
			queryParams: "?page=1&limit=10",
			setupClaims: func(ctx *gin.Context) {
				ctx.Set("claims", &jwt.CustomClaims{
					UserID: "user-id-123",
				})
			},
			setupMock: func(ctx context.Context, mockSvc *mocks.Service) {
				mockSvc.
					On(
						"List",
						ctx,
						"user-id-123",
						&bookmarkDTO.ListBookmarksRequest{
							Page:  1,
							Limit: 10,
							Sort:  "created_at",
						},
					).
					Return(
						nil,
						nil,
						errors.New("database error"),
					).
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
				http.MethodGet,
				"/v1/bookmarks"+tc.queryParams,
				nil,
			)

			ctx.Request = httpRequest

			tc.setupClaims(ctx)
			tc.setupMock(ctx, mockSvc)

			handler := NewHandler(mockSvc)
			handler.List(ctx)

			assert.Equal(t, tc.expected.statusCode, recorder.Code)
			assert.Equal(t, "application/json; charset=utf-8", recorder.Header().Get("Content-Type"))
			assert.Contains(t, recorder.Body.String(), tc.expected.bodyContains)

			mockSvc.AssertExpectations(t)
		})
	}
}
