package logger

import (
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// NewLoggerClient initializes the global logger with configuration loaded from environment variables
func NewLoggerClient(envPrefix string) error {
	config, err := LoadLoggerConfig(envPrefix)
	if err != nil {
		return err
	}

	level, err := zerolog.ParseLevel(strings.ToLower(config.Level))
	if err != nil {
		level = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(level)

	log.Logger = zerolog.New(os.Stdout).
		With().
		Timestamp().
		Logger()

	return nil
}
