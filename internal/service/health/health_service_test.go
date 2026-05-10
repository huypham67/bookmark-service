package health

import (
	"errors"
	"testing"

	response2 "github.com/huypham67/bookmark-management/internal/dto/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockPinger struct {
	mock.Mock
}

func (m *mockPinger) Ping() error {
	args := m.Called()
	return args.Error(0)
}

func TestHealthCheckService(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		serviceName    string
		instanceID     string
		setupPinger    func(*mockPinger)
		verifyResponse func(*testing.T, *response2.HealthCheckResponse, error) bool
	}{
		{
			name:        "should create service with provided values",
			serviceName: "bookmark-service",
			instanceID:  "instance-1",
			setupPinger: func(mp *mockPinger) {
				mp.On("Ping").Return(nil)
			},
			verifyResponse: func(t *testing.T, res *response2.HealthCheckResponse, err error) bool {
				return assert.NoError(t, err) &&
					assert.Equal(t, "OK", res.Message) &&
					assert.Equal(t, "bookmark-service", res.ServiceName) &&
					assert.Equal(t, "instance-1", res.InstanceID)
			},
		},
		{
			name:        "should create service with empty values",
			serviceName: "",
			instanceID:  "",
			setupPinger: func(mp *mockPinger) {
				mp.On("Ping").Return(nil)
			},
			verifyResponse: func(t *testing.T, res *response2.HealthCheckResponse, err error) bool {
				return assert.NoError(t, err) &&
					assert.Equal(t, "OK", res.Message) &&
					assert.Equal(t, "", res.ServiceName) &&
					assert.Equal(t, "", res.InstanceID)
			},
		},
		{
			name:        "should create service with only service name",
			serviceName: "test-service",
			instanceID:  "",
			setupPinger: func(mp *mockPinger) {
				mp.On("Ping").Return(nil)
			},
			verifyResponse: func(t *testing.T, res *response2.HealthCheckResponse, err error) bool {
				return assert.NoError(t, err) &&
					assert.Equal(t, "OK", res.Message) &&
					assert.Equal(t, "test-service", res.ServiceName)
			},
		},
		{
			name:        "should create service with only instance ID",
			serviceName: "",
			instanceID:  "test-instance",
			setupPinger: func(mp *mockPinger) {
				mp.On("Ping").Return(nil)
			},
			verifyResponse: func(t *testing.T, res *response2.HealthCheckResponse, err error) bool {
				return assert.NoError(t, err) &&
					assert.Equal(t, "OK", res.Message) &&
					assert.Equal(t, "test-instance", res.InstanceID)
			},
		},
		{
			name:        "should return FAILED message when ping fails",
			serviceName: "bookmark-service",
			instanceID:  "instance-1",
			setupPinger: func(mp *mockPinger) {
				mp.On("Ping").Return(errors.New("connection refused"))
			},
			verifyResponse: func(t *testing.T, res *response2.HealthCheckResponse, err error) bool {
				return assert.Error(t, err) &&
					assert.Equal(t, "FAILED", res.Message) &&
					assert.Equal(t, "bookmark-service", res.ServiceName) &&
					assert.Equal(t, "instance-1", res.InstanceID)
			},
		},
		{
			name:        "should return FAILED message with empty values on error",
			serviceName: "",
			instanceID:  "",
			setupPinger: func(mp *mockPinger) {
				mp.On("Ping").Return(errors.New("timeout"))
			},
			verifyResponse: func(t *testing.T, res *response2.HealthCheckResponse, err error) bool {
				return assert.Error(t, err) &&
					assert.Equal(t, "FAILED", res.Message) &&
					assert.Equal(t, "", res.ServiceName) &&
					assert.Equal(t, "", res.InstanceID)
			},
		},
		{
			name:        "should preserve service name and instance ID on error",
			serviceName: "critical-service",
			instanceID:  "prod-instance-1",
			setupPinger: func(mp *mockPinger) {
				mp.On("Ping").Return(errors.New("unreachable"))
			},
			verifyResponse: func(t *testing.T, res *response2.HealthCheckResponse, err error) bool {
				return assert.Error(t, err) &&
					assert.Equal(t, "FAILED", res.Message) &&
					assert.Equal(t, "critical-service", res.ServiceName) &&
					assert.Equal(t, "prod-instance-1", res.InstanceID)
			},
		},
		{
			name:        "should handle unknown error type",
			serviceName: "test-service",
			instanceID:  "test-instance",
			setupPinger: func(mp *mockPinger) {
				mp.On("Ping").Return(errors.New("unknown error"))
			},
			verifyResponse: func(t *testing.T, res *response2.HealthCheckResponse, err error) bool {
				return assert.Error(t, err) &&
					assert.Equal(t, "FAILED", res.Message) &&
					assert.Equal(t, "test-service", res.ServiceName) &&
					assert.Equal(t, "test-instance", res.InstanceID)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mp := &mockPinger{}
			tc.setupPinger(mp)

			service := NewHealthCheckService(tc.serviceName, tc.instanceID, mp)

			// Test GetStatus
			response, err := service.GetStatus()

			_ = tc.verifyResponse(t, &response, err)

			assert.Implements(t, (*HealthCheck)(nil), service)
			assert.NotNil(t, service)
		})
	}
}
