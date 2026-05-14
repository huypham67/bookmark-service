package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/huypham67/bookmark-management/internal/dto/response"
	"github.com/huypham67/bookmark-management/internal/service/mocks"
	"github.com/stretchr/testify/assert"
)

func TestHealthCheckHandler_GetHealthCheck(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		setupMock      func(*mocks.HealthCheckService)
		expectedCode   int
		verifyResponse func(*testing.T, *response.HealthCheckResponse) bool
	}{
		{
			name: "should return 200 with correct service name and instance ID",
			setupMock: func(mockService *mocks.HealthCheckService) {
				mockService.On("GetStatus").Return(
					response.HealthCheckResponse{
						Message:     "OK",
						ServiceName: "bookmark-service",
						InstanceID:  "instance-1",
					},
				)
			},
			expectedCode: http.StatusOK,
			verifyResponse: func(t *testing.T, res *response.HealthCheckResponse) bool {
				return assert.Equal(t, "OK", res.Message) &&
					assert.Equal(t, "bookmark-service", res.ServiceName) &&
					assert.Equal(t, "instance-1", res.InstanceID)
			},
		},
		{
			name: "should return 200 with different service configuration",
			setupMock: func(mockService *mocks.HealthCheckService) {
				mockService.On("GetStatus").Return(
					response.HealthCheckResponse{
						Message:     "OK",
						ServiceName: "auth-service",
						InstanceID:  "prod-2",
					},
				)
			},
			expectedCode: http.StatusOK,
			verifyResponse: func(t *testing.T, res *response.HealthCheckResponse) bool {
				return assert.Equal(t, "OK", res.Message) &&
					assert.Equal(t, "auth-service", res.ServiceName) &&
					assert.Equal(t, "prod-2", res.InstanceID)
			},
		},
		{
			name: "should return 500 when service returns error",
			setupMock: func(mockService *mocks.HealthCheckService) {
				mockService.On("GetStatus").Return(
					response.HealthCheckResponse{
						Message:     "FAILED",
						ServiceName: "bookmark-service",
						InstanceID:  "instance-1",
					},
				)
			},
			expectedCode: http.StatusInternalServerError,
			verifyResponse: func(t *testing.T, res *response.HealthCheckResponse) bool {
				return assert.Equal(t, "FAILED", res.Message) &&
					assert.Equal(t, "bookmark-service", res.ServiceName) &&
					assert.Equal(t, "instance-1", res.InstanceID)
			},
		},
		{
			name: "should return 500 with FAILED status on timeout error",
			setupMock: func(mockService *mocks.HealthCheckService) {
				mockService.On("GetStatus").Return(
					response.HealthCheckResponse{
						Message:     "FAILED",
						ServiceName: "test-service",
						InstanceID:  "prod-1",
					},
				)
			},
			expectedCode: http.StatusInternalServerError,
			verifyResponse: func(t *testing.T, res *response.HealthCheckResponse) bool {
				return assert.Equal(t, "FAILED", res.Message) &&
					assert.Equal(t, "test-service", res.ServiceName) &&
					assert.Equal(t, "prod-1", res.InstanceID)
			},
		},
		{
			name: "should handle empty service name on success",
			setupMock: func(mockService *mocks.HealthCheckService) {
				mockService.On("GetStatus").Return(
					response.HealthCheckResponse{
						Message:     "OK",
						ServiceName: "",
						InstanceID:  "test-instance",
					},
				)
			},
			expectedCode: http.StatusOK,
			verifyResponse: func(t *testing.T, res *response.HealthCheckResponse) bool {
				return assert.Equal(t, "OK", res.Message) &&
					assert.Equal(t, "", res.ServiceName) &&
					assert.Equal(t, "test-instance", res.InstanceID)
			},
		},
		{
			name: "should handle empty instance ID on success",
			setupMock: func(mockService *mocks.HealthCheckService) {
				mockService.On("GetStatus").Return(
					response.HealthCheckResponse{
						Message:     "OK",
						ServiceName: "my-service",
						InstanceID:  "",
					},
				)
			},
			expectedCode: http.StatusOK,
			verifyResponse: func(t *testing.T, res *response.HealthCheckResponse) bool {
				return assert.Equal(t, "OK", res.Message) &&
					assert.Equal(t, "my-service", res.ServiceName) &&
					assert.Equal(t, "", res.InstanceID)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gin.SetMode(gin.TestMode)

			mockService := mocks.NewHealthCheckService(t)
			tc.setupMock(mockService)

			handler := NewHealthCheckHandler(mockService)

			recorder := httptest.NewRecorder()

			ctx, _ := gin.CreateTestContext(recorder)

			request := httptest.NewRequest(
				http.MethodGet,
				"/health-check",
				nil,
			)

			ctx.Request = request

			handler.GetHealthCheck(ctx)

			assert.Equal(t, tc.expectedCode, recorder.Code)

			var actualResponse response.HealthCheckResponse

			err := json.Unmarshal(recorder.Body.Bytes(), &actualResponse)

			assert.NoError(t, err)

			_ = tc.verifyResponse(t, &actualResponse)

			mockService.AssertExpectations(t)
		})
	}
}
