package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/huypham67/bookmark-management/internal/model"
	"github.com/huypham67/bookmark-management/mocks"
	"github.com/stretchr/testify/assert"
)

func TestHealthCheckHandler_GetHealthCheck(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                string
		mockResponse        model.HealthCheckResponse
		expectedCode        int
		expectedServiceName string
		expectedInstanceID  string
	}{
		{
			name: "should return 200 with correct service name and instance ID",
			mockResponse: model.HealthCheckResponse{
				Message:     "OK",
				ServiceName: "bookmark-service",
				InstanceID:  "instance-1",
			},
			expectedCode:        http.StatusOK,
			expectedServiceName: "bookmark-service",
			expectedInstanceID:  "instance-1",
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gin.SetMode(gin.TestMode)

			mockService := mocks.NewHealthCheck(t)

			mockService.
				On("GetStatus").
				Return(tc.mockResponse)

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

			assert.Equal(
				t,
				tc.expectedCode,
				recorder.Code,
			)

			var actualResponse model.HealthCheckResponse

			err := json.Unmarshal(
				recorder.Body.Bytes(),
				&actualResponse,
			)

			assert.NoError(t, err)

			assert.Equal(
				t,
				tc.expectedServiceName,
				actualResponse.ServiceName,
			)

			assert.Equal(
				t,
				tc.expectedInstanceID,
				actualResponse.InstanceID,
			)

			mockService.AssertExpectations(t)
		})
	}
}
