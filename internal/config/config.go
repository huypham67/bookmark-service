package config

import (
	"strings"

	"github.com/google/uuid"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	AppPort     string `envconfig:"APP_PORT" default:"8080"`
	ServiceName string `envconfig:"SERVICE_NAME" required:"true"`
	InstanceID  string `envconfig:"INSTANCE_ID"`
}

// LoadConfig loads the configuration from environment variables and returns a Config struct.
func LoadConfig() (*Config, error) {
	config := &Config{}

	err := envconfig.Process("", config)
	if err != nil {
		return nil, err
	}

	config.ServiceName = strings.TrimSpace(config.ServiceName)
	config.InstanceID = strings.TrimSpace(config.InstanceID)

	if config.InstanceID == "" {
		config.InstanceID = uuid.New().String()
	}

	return config, nil
}
