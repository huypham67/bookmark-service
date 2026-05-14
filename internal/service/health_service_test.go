package service

import (
	"errors"
	"testing"

	"github.com/huypham67/bookmark-management/internal/dto/response"
	"github.com/huypham67/bookmark-management/pkg/redis"
	"github.com/stretchr/testify/assert"
)

func TestHealthCheckService_GetStatus(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		serviceName    string
		instanceID     string
		setupMock      func(*redis.MockPinger)
		verifyResponse func(*testing.T, response.HealthCheckResponse)
	}{
		{
			name:        "should return OK when ping succeeds",
			serviceName: "bookmark-service",
			instanceID:  "instance-1",
			setupMock: func(mp *redis.MockPinger) {
				mp.On("Ping").Return(nil).Once()
			},
			verifyResponse: func(t *testing.T, res response.HealthCheckResponse) {
				assert.Equal(t, response.HealthCheckResponse{
					Message:     statusMessage,
					ServiceName: "bookmark-service",
					InstanceID:  "instance-1",
				}, res)
			},
		},
		{
			name:        "should return FAILED when ping fails",
			serviceName: "bookmark-service",
			instanceID:  "instance-1",
			setupMock: func(mp *redis.MockPinger) {
				mp.On("Ping").
					Return(errors.New("redis connection failed")).
					Once()
			},
			verifyResponse: func(t *testing.T, res response.HealthCheckResponse) {
				assert.Equal(t, response.HealthCheckResponse{
					Message:     failedStatusMessage,
					ServiceName: "bookmark-service",
					InstanceID:  "instance-1",
				}, res)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			mp := &redis.MockPinger{}
			tc.setupMock(mp)

			service := NewHealthCheckService(
				tc.serviceName,
				tc.instanceID,
				mp,
			)

			// Act
			res := service.GetStatus()

			// Assert
			tc.verifyResponse(t, res)
			mp.AssertExpectations(t)
			mp.AssertNumberOfCalls(t, "Ping", 1)
		})
	}
}
