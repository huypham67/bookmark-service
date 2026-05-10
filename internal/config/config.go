package config

import (
	"strings"

	"github.com/google/uuid"
	"github.com/kelseyhightower/envconfig"
)

// AppConfig holds the application-specific configuration.
type AppConfig struct {
	AppPort     string `envconfig:"APP_PORT" default:"8080"`
	ServiceName string `envconfig:"SERVICE_NAME" required:"true"`
	InstanceID  string `envconfig:"INSTANCE_ID"`
}

// RedisConfig holds the Redis connection configuration.
type RedisConfig struct {
	Host     string `envconfig:"REDIS_HOST" default:"localhost"`
	Port     string `envconfig:"REDIS_PORT" default:"6379"`
	Password string `envconfig:"REDIS_PASSWORD" default:""`
	Database int    `envconfig:"REDIS_DATABASE" default:"0"`
}

// Config holds the application configuration loaded from environment variables.
type Config struct {
	*AppConfig
	*RedisConfig
}

// LoadConfig loads application configuration from environment variables.
func LoadConfig() (*Config, error) {
	appConfig := &AppConfig{}
	redisConfig := &RedisConfig{}

	err := envconfig.Process("", appConfig)
	if err != nil {
		return nil, err
	}

	err = envconfig.Process("", redisConfig)
	if err != nil {
		return nil, err
	}

	appConfig.ServiceName = strings.TrimSpace(appConfig.ServiceName)
	appConfig.InstanceID = strings.TrimSpace(appConfig.InstanceID)

	if appConfig.InstanceID == "" {
		appConfig.InstanceID = uuid.New().String()
	}

	return &Config{
		AppConfig:   appConfig,
		RedisConfig: redisConfig,
	}, nil
}
