package service

import (
	"context"
	"errors"
	"testing"

	"github.com/huypham67/bookmark-service/internal/dto/response"
	"github.com/huypham67/bookmark-service/internal/repository/mocks"
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
		setupMock      func(context.Context, *mocks.MockPinger)
		verifyResponse func(*testing.T, response.HealthCheckResponse)
	}{
		{
			name: "should return OK when redis ping succeeds",
			fields: fields{
				serviceName: "bookmark-service",
				instanceID:  "instance-1",
			},
			setupMock: func(ctx context.Context, mp *mocks.MockPinger) {
				mp.On("Ping", ctx).Return(nil).Once()
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
			setupMock: func(ctx context.Context, mp *mocks.MockPinger) {
				mp.On("Ping", ctx).
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
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			mp := &mocks.MockPinger{}
			tc.setupMock(ctx, mp)

			service := NewHealthCheckService(
				tc.fields.serviceName,
				tc.fields.instanceID,
				mp,
			)

			res := service.GetStatus(ctx)

			tc.verifyResponse(t, res)
			mp.AssertExpectations(t)
			mp.AssertNumberOfCalls(t, "Ping", 1)
		})
	}
}
