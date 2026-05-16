package logger

import "github.com/kelseyhightower/envconfig"

// LoggerConfig holds the configuration for the logger.
type LoggerConfig struct {
	Level string `envconfig:"LOG_LEVEL" default:"info"`
}

// LoadLoggerConfig loads logger configuration from environment variables with the given prefix.
func LoadLoggerConfig(prefix string) (*LoggerConfig, error) {
	cfg := &LoggerConfig{}

	if err := envconfig.Process(prefix, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
