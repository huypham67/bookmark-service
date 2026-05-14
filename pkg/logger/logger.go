package logger

import (
	"os"
	"strings"

	"github.com/rs/zerolog"
)

var globalLogger *zerolog.Logger

// Init initializes the global logger with configuration loaded from environment variables
func Init(envPrefix string) error {
	config, err := LoadLoggerConfig(envPrefix)
	if err != nil {
		return err
	}

	level, err := zerolog.ParseLevel(strings.ToLower(config.Level))
	if err != nil {
		level = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(level)

	logger := zerolog.New(os.Stdout).
		With().
		Timestamp().
		Logger()
	globalLogger = &logger

	return nil
}

// Get returns the global logger instance
func Get() *zerolog.Logger {
	return globalLogger
}
