package health

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	healthDTO "github.com/huypham67/bookmark-service/internal/dto/health"
	"github.com/huypham67/bookmark-service/internal/service/health/mocks"
	"github.com/stretchr/testify/assert"
)

func TestHandler_GetHealthCheck(t *testing.T) {
	t.Parallel()

	type expected struct {
		statusCode int
		response   healthDTO.HealthCheckResponse
	}

	testCases := []struct {
		name      string
		setupMock func(context.Context, *mocks.Service)
		expected  expected
	}{
		{
			name: "should return 200 OK when health check is successful",
			setupMock: func(ctx context.Context, m *mocks.Service) {
				m.On("GetStatus", ctx).Return(healthDTO.HealthCheckResponse{
					Message:     "OK",
					ServiceName: "bookmark-service",
					InstanceID:  "instance-1",
				}).Once()
			},
			expected: expected{
				statusCode: http.StatusOK,
				response: healthDTO.HealthCheckResponse{
					Message:     "OK",
					ServiceName: "bookmark-service",
					InstanceID:  "instance-1",
				},
			},
		}, {
			name: "should return 500 when health check is failed",
			setupMock: func(ctx context.Context, m *mocks.Service) {
				m.On("GetStatus", ctx).Return(healthDTO.HealthCheckResponse{Message: "FAILED",
					ServiceName: "bookmark-service",
					InstanceID:  "instance-1",
				})
			},
			expected: expected{
				statusCode: http.StatusInternalServerError,
				response: healthDTO.HealthCheckResponse{
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
			mockSvc := mocks.NewService(t)

			handler := NewHandler(mockSvc)
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			httpRequest := httptest.NewRequest(http.MethodGet, "/health-check", nil)
			ctx.Request = httpRequest

			tc.setupMock(ctx, mockSvc)
			handler.GetHealthCheck(ctx)

			assert.Equal(t, tc.expected.statusCode, recorder.Code)
			assert.Equal(t, "application/json; charset=utf-8", recorder.Header().Get("Content-Type"))

			var actual healthDTO.HealthCheckResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &actual)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected.response, actual)

			mockSvc.AssertExpectations(t)
		})
	}

}
