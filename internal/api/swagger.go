package api

import (
	"strings"

	"github.com/huypham67/bookmark-service/docs"
)

// SetupSwagger configures Swagger documentation with provided schemes and base path.
// This encapsulates all Swagger doc configuration in one place.
func SetupSwagger(schemes []string, basePath string) {
	docs.SwaggerInfo.Host = ""
	docs.SwaggerInfo.Schemes = schemes
	docs.SwaggerInfo.BasePath = basePath
}

// DetermineSchemes returns appropriate Swagger schemes based on environment and custom configuration.
// Priority: custom schemes > production default (https) > development default (http)
func DetermineSchemes(environment string, customSchemes string) []string {
	if customSchemes != "" {
		return parseSchemes(customSchemes)
	}

	if environment == "production" {
		return []string{"https"}
	}

	return []string{"http"}
}

// parseSchemes parses a comma-separated scheme string into a slice.
// Example: "http,https" → []string{"http", "https"}
func parseSchemes(schemesStr string) []string {
	schemes := make([]string, 0)
	for _, scheme := range strings.Split(schemesStr, ",") {
		if trimmed := strings.TrimSpace(scheme); trimmed != "" {
			schemes = append(schemes, trimmed)
		}
	}
	return schemes
}
