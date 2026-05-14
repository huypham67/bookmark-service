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

	type fields struct {
		serviceName string
		instanceID  string
	}

	testCases := []struct {
		name           string
		fields         fields
		setupMock      func(*redis.MockPinger)
		verifyResponse func(*testing.T, response.HealthCheckResponse)
	}{
		{
			name: "should return OK when redis ping succeeds",
			fields: fields{
				serviceName: "bookmark-service",
				instanceID:  "instance-1",
			},
			setupMock: func(mp *redis.MockPinger) {
				mp.On("Ping").Return(nil).Once()
			},
			verifyResponse: func(t *testing.T, res response.HealthCheckResponse) {
				assert.Equal(t, statusMessage, res.Message)
				assert.Equal(t, "bookmark-service", res.ServiceName)
				assert.Equal(t, "instance-1", res.InstanceID)
			},
		},
		{
			name: "should return FAILED when ping fails",
			fields: fields{
				serviceName: "bookmark-service",
				instanceID:  "instance-1",
			},
			setupMock: func(mp *redis.MockPinger) {
				mp.On("Ping").
					Return(errors.New("redis connection failed")).
					Once()
			},
			verifyResponse: func(t *testing.T, res response.HealthCheckResponse) {
				assert.Equal(t, failedStatusMessage, res.Message)
				assert.Equal(t, "bookmark-service", res.ServiceName)
				assert.Equal(t, "instance-1", res.InstanceID)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mp := &redis.MockPinger{}
			tc.setupMock(mp)

			service := NewHealthCheckService(
				tc.fields.serviceName,
				tc.fields.instanceID,
				mp,
			)

			res := service.GetStatus()

			tc.verifyResponse(t, res)
			mp.AssertExpectations(t)
			mp.AssertNumberOfCalls(t, "Ping", 1)
		})
	}
}
