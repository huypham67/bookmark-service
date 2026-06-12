package link

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	linkDTO "github.com/huypham67/bookmark-service/internal/dto/link"
	"github.com/huypham67/bookmark-service/internal/service/link/mocks"
	"github.com/stretchr/testify/assert"
)

func TestHandler_ShortenURL(t *testing.T) {
	t.Parallel()

	type expected struct {
		statusCode int
		body       string
	}

	testCases := []struct {
		name        string
		requestBody string
		setupMock   func(context.Context, *mocks.Service)
		expected    expected
	}{
		{
			name: "should return 200 when shorten URL succeeds",
			requestBody: `{
				"url": "https://www.google.com",
				"exp": 3600
			}`,
			setupMock: func(ctx context.Context, mockService *mocks.Service) {
				mockService.
					On(
						"ShortenURL",
						ctx,
						linkDTO.ShortenURLRequest{
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
			setupMock: func(ctx context.Context, mockService *mocks.Service) {
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
			setupMock: func(ctx context.Context, mockService *mocks.Service) {
				mockService.
					On(
						"ShortenURL",
						ctx,
						linkDTO.ShortenURLRequest{
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

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gin.SetMode(gin.TestMode)
			mockSvc := new(mocks.Service)

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

			handler := NewHandler(mockSvc)

			handler.ShortenURL(ctx)

			assert.Equal(t, tc.expected.statusCode, recorder.Code)
			assert.Equal(t, "application/json; charset=utf-8", recorder.Header().Get("Content-Type"))
			assert.Contains(t, recorder.Body.String(), tc.expected.body)

			mockSvc.AssertExpectations(t)
		})
	}

}
