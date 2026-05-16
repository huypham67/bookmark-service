package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/huypham67/bookmark-service/internal/dto/response"
	"github.com/huypham67/bookmark-service/internal/service/mocks"
	"github.com/stretchr/testify/assert"
)

func TestHealthCheckHandler_GetHealthCheck(t *testing.T) {
	t.Parallel()

	type expected struct {
		statusCode int
		response   response.HealthCheckResponse
	}

	testCases := []struct {
		name      string
		setupMock func(context.Context, *mocks.HealthCheckService)
		expected  expected
	}{
		{
			name: "should return 200 OK when health check is successful",
			setupMock: func(ctx context.Context, m *mocks.HealthCheckService) {
				m.On("GetStatus", ctx).Return(response.HealthCheckResponse{
					Message:     "OK",
					ServiceName: "bookmark-service",
					InstanceID:  "instance-1",
				}).Once()
			},
			expected: expected{
				statusCode: http.StatusOK,
				response: response.HealthCheckResponse{
					Message:     "OK",
					ServiceName: "bookmark-service",
					InstanceID:  "instance-1",
				},
			},
		}, {
			name: "should return 500 when health check is failed",
			setupMock: func(ctx context.Context, m *mocks.HealthCheckService) {
				m.On("GetStatus", ctx).Return(response.HealthCheckResponse{Message: "FAILED",
					ServiceName: "bookmark-service",
					InstanceID:  "instance-1",
				})
			},
			expected: expected{
				statusCode: http.StatusInternalServerError,
				response: response.HealthCheckResponse{
					Message:     "FAILED",
					ServiceName: "bookmark-service",
					InstanceID:  "instance-1",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gin.SetMode(gin.TestMode)
			mockSvc := mocks.NewHealthCheckService(t)

			handler := NewHealthCheckHandler(mockSvc)
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			httpRequest := httptest.NewRequest(http.MethodGet, "/health-check", nil)
			ctx.Request = httpRequest

			tc.setupMock(ctx, mockSvc)
			handler.GetHealthCheck(ctx)

			assert.Equal(t, tc.expected.statusCode, recorder.Code)
			assert.Equal(t, "application/json; charset=utf-8", recorder.Header().Get("Content-Type"))

			var actual response.HealthCheckResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &actual)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected.response, actual)

			mockSvc.AssertExpectations(t)
		})
	}

}
