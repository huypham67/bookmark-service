package config

import (
	"strings"

	"github.com/google/uuid"
	"github.com/kelseyhightower/envconfig"
)

// Config holds the application configuration loaded from environment variables.
type Config struct {
	AppPort     string `envconfig:"APP_PORT" default:"8080"`
	ServiceName string `envconfig:"SERVICE_NAME" required:"true"`
	InstanceID  string `envconfig:"INSTANCE_ID"`
	HostName    string `envconfig:"APP_HOST_NAME" default:"/api/bookmark_service"`
}

// LoadConfig loads application configuration from environment variables.
func LoadConfig() (*Config, error) {
	cfg := &Config{}

	err := envconfig.Process("", cfg)
	if err != nil {
		return nil, err
	}

	cfg.ServiceName = strings.TrimSpace(cfg.ServiceName)
	cfg.InstanceID = strings.TrimSpace(cfg.InstanceID)
	cfg.HostName = strings.TrimSpace(cfg.HostName)

	if cfg.InstanceID == "" {
		cfg.InstanceID = uuid.New().String()
	}

	return cfg, nil
}
