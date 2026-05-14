package main

import (
	"fmt"

	"github.com/huypham67/bookmark-management/pkg/logger"
	"github.com/rs/zerolog/log"
)

func main() {
	// Initialize logger with configuration from environment variables with "LOG" prefix
	// Environment variable: LOG_LEVEL (default: "info")
	err := logger.Init("LOG")
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize logger")
		return
	}

	// Display the loaded log level
	cfg, err := logger.LoadLoggerConfig("LOG")
	if err != nil {
		log.Error().Err(err).Msg("Failed to load logger config")
		return
	}
	fmt.Printf("Current Log Level: %s\n", cfg.Level)

	// Test logging at different levels
	log.Debug().Msg("This is a debug message")
	log.Info().Msg("This is an info message")
	log.Warn().Msg("This is a warning message")
	log.Error().Msg("This is an error message")

	fmt.Println("\nLogger initialized successfully with level:", cfg.Level)
}
