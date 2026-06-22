package bootstrap

import (
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/huypham67/bookmark-common/pkg/logger"
	"github.com/huypham67/bookmark-service/internal/api"
)

// App represents the application and manages its runtime lifecycle.
type App struct {
	container *Container
	router    *api.Router
}

// NewApp initializes the application by setting up logging, configuration, dependencies, and routing.
func NewApp() (*App, error) {
	if err := logger.NewClient(""); err != nil {
		return nil, err
	}

	container, err := NewContainer()
	if err != nil {
		log.Error().Err(err).Msg("failed to initialize container")
		return nil, err
	}

	// Setup Swagger configuration
	schemes := api.DetermineSchemes(container.Config.Environment, container.Config.SwaggerSchemes)
	api.SetupSwagger(schemes, container.Config.HostName)

	// Create router
	router := api.NewRouter()

	// Register all routes using Container dependencies
	SetupRoutes(router, container)

	log.Info().Msg("application initialized successfully")

	return &App{
		container: container,
		router:    router,
	}, nil
}

// Run starts the HTTP server on the configured port.
func (a *App) Run() error {
	return http.ListenAndServe(
		":"+a.container.Config.AppPort,
		a.router,
	)
}

// Close gracefully shuts down all application resources.
func (a *App) Close() error {
	return a.container.Close()
}
