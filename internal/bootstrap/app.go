package bootstrap

import (
	"github.com/huypham67/bookmark-management/internal/api"
	"github.com/huypham67/bookmark-management/internal/config"
	"github.com/huypham67/bookmark-management/internal/handler"
	"github.com/huypham67/bookmark-management/internal/service"
)

type App struct {
	router *api.Router
}

func NewApp() (*App, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	healthService := service.NewHealthCheckService(
		cfg.ServiceName,
		cfg.InstanceID,
	)

	healthHandler := handler.NewHealthCheckHandler(
		healthService,
	)

	router := api.NewRouter(
		cfg.AppPort,
		healthHandler,
	)

	return &App{
		router: router,
	}, nil
}

func (a *App) Run() error {
	return a.router.Run()
}
