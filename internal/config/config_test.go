package config

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {

	testCases := []struct {
		name                    string
		setupEnv                func()
		cleanupEnv              func()
		expectError             bool
		expectedServiceName     string
		expectGeneratedInstance bool
	}{
		{
			name: "should load config with all required values provided",
			setupEnv: func() {
				t.Setenv("SERVICE_NAME", "bookmark-service")
				t.Setenv("INSTANCE_ID", "instance-1")
				t.Setenv("APP_PORT", "9090")
			},
			cleanupEnv:              func() {},
			expectError:             false,
			expectedServiceName:     "bookmark-service",
			expectGeneratedInstance: false,
		},
		{
			name: "should use default APP_PORT when not provided",
			setupEnv: func() {
				t.Setenv("SERVICE_NAME", "test-service")
				t.Setenv("INSTANCE_ID", "")
			},
			cleanupEnv:              func() {},
			expectError:             false,
			expectedServiceName:     "test-service",
			expectGeneratedInstance: true,
		},
		{
			name: "should generate UUID for InstanceID when empty",
			setupEnv: func() {
				t.Setenv("SERVICE_NAME", "my-service")
				t.Setenv("INSTANCE_ID", "")
			},
			cleanupEnv:              func() {},
			expectError:             false,
			expectedServiceName:     "my-service",
			expectGeneratedInstance: true,
		},
		{
			name: "should trim whitespace from ServiceName and InstanceID",
			setupEnv: func() {
				t.Setenv("SERVICE_NAME", "  booking-service  ")
				t.Setenv("INSTANCE_ID", "  instance-1  ")
			},
			cleanupEnv:              func() {},
			expectError:             false,
			expectedServiceName:     "booking-service",
			expectGeneratedInstance: false,
		},
		{
			name: "should handle empty ServiceName (no error from envconfig)",
			setupEnv: func() {
				t.Setenv("SERVICE_NAME", "")
				t.Setenv("INSTANCE_ID", "instance-1")
			},
			cleanupEnv:              func() {},
			expectError:             false,
			expectedServiceName:     "",
			expectGeneratedInstance: false,
		},
		{
			name: "should fail when SERVICE_NAME is not set at all",
			setupEnv: func() {
				t.Setenv("SERVICE_NAME", "")
				t.Setenv("INSTANCE_ID", "instance-1")
			},
			cleanupEnv:              func() {},
			expectError:             false,
			expectedServiceName:     "",
			expectGeneratedInstance: false,
		},
		{
			name: "should generate UUID for InstanceID when not set at all",
			setupEnv: func() {
				t.Setenv("SERVICE_NAME", "api-service")
				t.Setenv("INSTANCE_ID", "")
			},
			cleanupEnv:              func() {},
			expectError:             false,
			expectedServiceName:     "api-service",
			expectGeneratedInstance: true,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {

			tc.setupEnv()
			defer tc.cleanupEnv()

			config, err := LoadConfig()

			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, config)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, config)
			assert.NotNil(t, config.AppConfig)
			assert.NotNil(t, config.RedisConfig)
			assert.Equal(t, tc.expectedServiceName, config.AppConfig.ServiceName)

			if tc.expectGeneratedInstance {
				assert.NotEmpty(t, config.AppConfig.InstanceID)
				_, err := uuid.Parse(config.AppConfig.InstanceID)
				assert.NoError(t, err, "InstanceID should be a valid UUID")
			} else {
				assert.NotEmpty(t, config.AppConfig.InstanceID)
			}

			// Verify default APP_PORT
			if config.AppConfig.AppPort == "" {
				assert.Equal(t, "8080", config.AppConfig.AppPort)
			}
		})
	}
}
