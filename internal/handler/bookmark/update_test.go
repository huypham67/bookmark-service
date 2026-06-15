package bookmark

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/huypham67/bookmark-common/pkg/jwt"
	bookmarkDTO "github.com/huypham67/bookmark-service/internal/dto/bookmark"
	"github.com/huypham67/bookmark-service/internal/service/bookmark"
	"github.com/huypham67/bookmark-service/internal/service/bookmark/mocks"
	"github.com/stretchr/testify/assert"
)

func TestHandler_Update(t *testing.T) {
	t.Parallel()

	type expected struct {
		statusCode   int
		bodyContains string
	}

	testCases := []struct {
		name        string
		bookmarkID  string
		requestBody string
		setupClaims func(*gin.Context)
		setupMock   func(context.Context, *mocks.Service)
		expected    expected
	}{
		{
			name:       "should return 200 when bookmark is updated successfully",
			bookmarkID: "bm-123",
			requestBody: `{
				"description":"Updated Description",
				"url":"https://updated.com"
			}`,
			setupClaims: func(ctx *gin.Context) {
				ctx.Set("claims", &jwt.CustomClaims{
					UserID: "user-id-123",
				})
			},
			setupMock: func(ctx context.Context, mockSvc *mocks.Service) {
				mockSvc.
					On(
						"Update",
						ctx,
						"user-id-123",
						"bm-123",
						bookmarkDTO.UpdateBookmarkRequest{
							ID:          "bm-123",
							Description: "Updated Description",
							URL:         "https://updated.com",
						},
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
			name:        "should return 401 when user ID is not found in context",
			bookmarkID:  "bm-123",
			requestBody: `{"description":"Updated Description"}`,
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
			name:        "should return 400 when request body is invalid JSON",
			bookmarkID:  "bm-123",
			requestBody: `{invalid json}`,
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
				bodyContains: "Invalid request",
			},
		},
		{
			name:       "should return 404 when bookmark is not found",
			bookmarkID: "nonexistent-bm",
			requestBody: `{
				"description":"Updated Description"
			}`,
			setupClaims: func(ctx *gin.Context) {
				ctx.Set("claims", &jwt.CustomClaims{
					UserID: "user-id-123",
				})
			},
			setupMock: func(ctx context.Context, mockSvc *mocks.Service) {
				mockSvc.
					On(
						"Update",
						ctx,
						"user-id-123",
						"nonexistent-bm",
						bookmarkDTO.UpdateBookmarkRequest{
							ID:          "nonexistent-bm",
							Description: "Updated Description",
							URL:         "",
						},
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
			requestBody: `{
				"description":"Updated Description",
				"url":"https://updated.com"
			}`,
			setupClaims: func(ctx *gin.Context) {
				ctx.Set("claims", &jwt.CustomClaims{
					UserID: "user-id-123",
				})
			},
			setupMock: func(ctx context.Context, mockSvc *mocks.Service) {
				mockSvc.
					On(
						"Update",
						ctx,
						"user-id-123",
						"bm-123",
						bookmarkDTO.UpdateBookmarkRequest{
							ID:          "bm-123",
							Description: "Updated Description",
							URL:         "https://updated.com",
						},
					).
					Return(bookmark.ErrBadRequest).
					Once()
			},
			expected: expected{
				statusCode:   http.StatusBadRequest,
				bodyContains: "Invalid request",
			},
		},
		{
			name:       "should return 500 when service returns unexpected error",
			bookmarkID: "bm-123",
			requestBody: `{
				"description":"Updated Description"
			}`,
			setupClaims: func(ctx *gin.Context) {
				ctx.Set("claims", &jwt.CustomClaims{
					UserID: "user-id-123",
				})
			},
			setupMock: func(ctx context.Context, mockSvc *mocks.Service) {
				mockSvc.
					On(
						"Update",
						ctx,
						"user-id-123",
						"bm-123",
						bookmarkDTO.UpdateBookmarkRequest{
							ID:          "bm-123",
							Description: "Updated Description",
							URL:         "",
						},
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
				http.MethodPut,
				"/v1/bookmarks/"+tc.bookmarkID,
				bytes.NewBufferString(tc.requestBody),
			)

			httpRequest.Header.Set("Content-Type", "application/json")
			ctx.Request = httpRequest
			ctx.Params = []gin.Param{{Key: "id", Value: tc.bookmarkID}}

			tc.setupClaims(ctx)
			tc.setupMock(ctx, mockSvc)

			handler := NewHandler(mockSvc, nil)
			handler.Update(ctx)

			assert.Equal(t, tc.expected.statusCode, recorder.Code)
			assert.Equal(t, "application/json; charset=utf-8", recorder.Header().Get("Content-Type"))
			assert.Contains(t, recorder.Body.String(), tc.expected.bodyContains)

			mockSvc.AssertExpectations(t)
		})
	}
}
