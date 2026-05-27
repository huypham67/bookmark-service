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
	"github.com/huypham67/bookmark-service/internal/model"
	"github.com/huypham67/bookmark-service/internal/service"
	serviceMocks "github.com/huypham67/bookmark-service/internal/service/mocks"
	"github.com/stretchr/testify/assert"
)

func TestUserHandler_Register(t *testing.T) {
	t.Parallel()

	type expected struct {
		statusCode   int
		bodyContains string
	}

	testCases := []struct {
		name        string
		requestBody string
		setupMock   func(context.Context, *serviceMocks.User)
		expected    expected
	}{
		{
			name: "should return 201 when register succeeds",
			requestBody: `{
				"display_name": "Test User",
				"username": "testuser",
				"email": "test@example.com",
				"password": "password123"
			}`,
			setupMock: func(ctx context.Context, mockService *serviceMocks.User) {
				mockService.
					On(
						"RegisterUser",
						ctx,
						request.RegisterUserRequest{
							DisplayName: "Test User",
							Username:    "testuser",
							Email:       "test@example.com",
							Password:    "password123",
						},
					).
					Return(&model.User{
						ID:          "uuid-123",
						DisplayName: "Test User",
						Username:    "testuser",
						Email:       "test@example.com",
					}, nil).
					Once()
			},
			expected: expected{
				statusCode:   http.StatusCreated,
				bodyContains: "Register an user successfully!",
			},
		},
		{
			name: "should return 400 when validation fails",
			requestBody: `{
				"display_name": "Test User"
			}`,
			setupMock: func(ctx context.Context, mockService *serviceMocks.User) {
				// No need to set up mock for validation failure
			},
			expected: expected{
				statusCode:   http.StatusBadRequest,
				bodyContains: "Invalid request body",
			},
		},
		{
			name: "should return 409 when email already registered",
			requestBody: `{
				"display_name": "Test User",
				"username": "testuser",
				"email": "existing@example.com",
				"password": "password123"
			}`,
			setupMock: func(ctx context.Context, mockService *serviceMocks.User) {
				mockService.
					On(
						"RegisterUser",
						ctx,
						request.RegisterUserRequest{
							DisplayName: "Test User",
							Username:    "testuser",
							Email:       "existing@example.com",
							Password:    "password123",
						},
					).
					Return(nil, service.ErrEmailAlreadyRegistered).
					Once()
			},
			expected: expected{
				statusCode:   http.StatusConflict,
				bodyContains: "User already exists",
			},
		},
		{
			name: "should return 409 when username already exists",
			requestBody: `{
				"display_name": "Test User",
				"username": "existinguser",
				"email": "test@example.com",
				"password": "password123"
			}`,
			setupMock: func(ctx context.Context, mockService *serviceMocks.User) {
				mockService.
					On(
						"RegisterUser",
						ctx,
						request.RegisterUserRequest{
							DisplayName: "Test User",
							Username:    "existinguser",
							Email:       "test@example.com",
							Password:    "password123",
						},
					).
					Return(nil, service.ErrUsernameAlreadyExists).
					Once()
			},
			expected: expected{
				statusCode:   http.StatusConflict,
				bodyContains: "User already exists",
			},
		},
		{
			name: "should return 500 when service returns error",
			requestBody: `{
				"display_name": "Test User",
				"username": "testuser",
				"email": "test@example.com",
				"password": "password123"
			}`,
			setupMock: func(ctx context.Context, mockService *serviceMocks.User) {
				mockService.
					On(
						"RegisterUser",
						ctx,
						request.RegisterUserRequest{
							DisplayName: "Test User",
							Username:    "testuser",
							Email:       "test@example.com",
							Password:    "password123",
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
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gin.SetMode(gin.TestMode)
			mockSvc := serviceMocks.NewUser(t)
			recorder := httptest.NewRecorder()

			ctx, _ := gin.CreateTestContext(recorder)

			httpRequest := httptest.NewRequest(
				http.MethodPost,
				"/v1/users/register",
				strings.NewReader(tc.requestBody),
			)

			httpRequest.Header.Set("Content-Type", "application/json")
			ctx.Request = httpRequest

			tc.setupMock(ctx, mockSvc)

			handler := NewUserHandler(mockSvc)

			handler.Register(ctx)

			assert.Equal(t, tc.expected.statusCode, recorder.Code)
			assert.Equal(t, "application/json; charset=utf-8", recorder.Header().Get("Content-Type"))
			assert.Contains(t, recorder.Body.String(), tc.expected.bodyContains)

			mockSvc.AssertExpectations(t)
		})
	}

}
