package bookmark

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/huypham67/bookmark-common/pkg/jwt"
	bookmarkDTO "github.com/huypham67/bookmark-service/internal/dto/bookmark"
	"github.com/huypham67/bookmark-service/internal/model"
	"github.com/huypham67/bookmark-service/internal/service/bookmark"
	"github.com/huypham67/bookmark-service/internal/service/bookmark/mocks"
	"github.com/stretchr/testify/assert"
)

func TestHandler_Create(t *testing.T) {
	t.Parallel()

	type expected struct {
		statusCode   int
		bodyContains string
	}

	testCases := []struct {
		name        string
		requestBody string
		setupClaims func(*gin.Context)
		setupMock   func(context.Context, *mocks.Service)
		expected    expected
	}{
		{
			name: "should return 201 when bookmark is created successfully",
			requestBody: `{
				"description":"Test Bookmark",
				"url":"https://example.com"
			}`,
			setupClaims: func(ctx *gin.Context) {
				ctx.Set("claims", &jwt.CustomClaims{
					UserID: "user-id-123",
				})
			},
			setupMock: func(ctx context.Context, mockSvc *mocks.Service) {
				mockSvc.
					On(
						"Create",
						ctx,
						"user-id-123",
						bookmarkDTO.CreateBookmarkRequest{
							Description: "Test Bookmark",
							URL:         "https://example.com",
						},
					).
					Return(
						&model.Bookmark{
							BaseModel: model.BaseModel{
								ID: "bm-123",
							},
							Description: "Test Bookmark",
							URL:         "https://example.com",
							Code:        "abc123",
							UserID:      "user-id-123",
						},
						nil,
					).
					Once()
			},
			expected: expected{
				statusCode:   http.StatusCreated,
				bodyContains: "Bookmark created successfully!",
			},
		},
		{
			name: "should return 401 when user ID is not found in context",
			requestBody: `{
				"description":"Test Bookmark",
				"url":"https://example.com"
			}`,
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
				bodyContains: "Invalid request body",
			},
		},
		{
			name: "should return 409 when bookmark code already exists",
			requestBody: `{
				"description":"Test Bookmark",
				"url":"https://example.com"
			}`,
			setupClaims: func(ctx *gin.Context) {
				ctx.Set("claims", &jwt.CustomClaims{
					UserID: "user-id-123",
				})
			},
			setupMock: func(ctx context.Context, mockSvc *mocks.Service) {
				mockSvc.
					On(
						"Create",
						ctx,
						"user-id-123",
						bookmarkDTO.CreateBookmarkRequest{
							Description: "Test Bookmark",
							URL:         "https://example.com",
						},
					).
					Return(nil, bookmark.ErrBookmarkAlreadyExists).
					Once()
			},
			expected: expected{
				statusCode:   http.StatusConflict,
				bodyContains: "Bookmark code already exists",
			},
		},
		{
			name: "should return 400 when service returns bad request error",
			requestBody: `{
				"description":"Test Bookmark",
				"url":"https://example.com"
			}`,
			setupClaims: func(ctx *gin.Context) {
				ctx.Set("claims", &jwt.CustomClaims{
					UserID: "nonexistent-user",
				})
			},
			setupMock: func(ctx context.Context, mockSvc *mocks.Service) {
				mockSvc.
					On(
						"Create",
						ctx,
						"nonexistent-user",
						bookmarkDTO.CreateBookmarkRequest{
							Description: "Test Bookmark",
							URL:         "https://example.com",
						},
					).
					Return(nil, bookmark.ErrBadRequest).
					Once()
			},
			expected: expected{
				statusCode:   http.StatusBadRequest,
				bodyContains: "Invalid bookmark request",
			},
		},
		{
			name: "should return 500 when service returns unexpected error",
			requestBody: `{
				"description":"Test Bookmark",
				"url":"https://example.com"
			}`,
			setupClaims: func(ctx *gin.Context) {
				ctx.Set("claims", &jwt.CustomClaims{
					UserID: "user-id-123",
				})
			},
			setupMock: func(ctx context.Context, mockSvc *mocks.Service) {
				mockSvc.
					On(
						"Create",
						ctx,
						"user-id-123",
						bookmarkDTO.CreateBookmarkRequest{
							Description: "Test Bookmark",
							URL:         "https://example.com",
						},
					).
					Return(nil, errors.New("database error")).
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
				http.MethodPost,
				"/v1/bookmarks",
				strings.NewReader(tc.requestBody),
			)

			httpRequest.Header.Set("Content-Type", "application/json")
			ctx.Request = httpRequest

			tc.setupClaims(ctx)
			tc.setupMock(ctx, mockSvc)

			handler := NewHandler(mockSvc)
			handler.Create(ctx)

			assert.Equal(t, tc.expected.statusCode, recorder.Code)
			assert.Equal(t, "application/json; charset=utf-8", recorder.Header().Get("Content-Type"))
			assert.Contains(t, recorder.Body.String(), tc.expected.bodyContains)

			mockSvc.AssertExpectations(t)
		})
	}
}
