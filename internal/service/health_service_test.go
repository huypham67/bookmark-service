package service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/huypham67/bookmark-management/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestHealthCheckService_GetStatus(t *testing.T) {
	testCases := []struct {
		name                    string
		serviceName             string
		instanceID              string
		expectedServiceName     string
		expectGeneratedInstance bool
	}{
		{
			name:                    "Valid: ServiceName and InstanceID are provided -> should return provided values",
			serviceName:             "bookmark-service",
			instanceID:              "instance-1",
			expectedServiceName:     "bookmark-service",
			expectGeneratedInstance: false,
		},
		{
			name:                    "should return provided service name and generated instance ID if instance ID is empty",
			serviceName:             "bookmark-service",
			instanceID:              "",
			expectedServiceName:     "bookmark-service",
			expectGeneratedInstance: true,
		},
		{
			name:                    "should return empty service name and provided instance ID if service name is empty",
			serviceName:             "",
			instanceID:              "instance-1",
			expectedServiceName:     "",
			expectGeneratedInstance: false,
		},
		{
			name:                    "should return empty service name and generated instance ID if service name and instance ID are empty",
			serviceName:             "",
			instanceID:              "",
			expectedServiceName:     "",
			expectGeneratedInstance: true,
		},
		{
			name:                    "should return generated instance ID if instance ID is not provided",
			serviceName:             "bookmark-service",
			expectedServiceName:     "bookmark-service",
			expectGeneratedInstance: true,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("SERVICE_NAME", tc.serviceName)
			t.Setenv("INSTANCE_ID", tc.instanceID)

			cfg, err := config.LoadConfig()

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedServiceName, cfg.ServiceName)

			if tc.expectGeneratedInstance {
				assert.NotEmpty(t, cfg.InstanceID)

				_, err := uuid.Parse(cfg.InstanceID)
				assert.NoError(t, err)
			} else {
				assert.Equal(t, tc.instanceID, cfg.InstanceID)
			}
		})
	}
}
