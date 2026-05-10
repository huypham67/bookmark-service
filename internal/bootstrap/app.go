package bootstrap

import (
	"github.com/huypham67/bookmark-management/infrastructure/redis"
	"github.com/huypham67/bookmark-management/internal/api"
	"github.com/huypham67/bookmark-management/internal/config"
	healthHandler "github.com/huypham67/bookmark-management/internal/handler/health"
	linkHandler "github.com/huypham67/bookmark-management/internal/handler/link"
	"github.com/huypham67/bookmark-management/internal/repository"
	healthService "github.com/huypham67/bookmark-management/internal/service/health"
	linkService "github.com/huypham67/bookmark-management/internal/service/link"
)

// App represents the application container holding the router.
type App struct {
	router *api.Router
}

// NewApp initializes and returns a new application instance with all dependencies configured.
func NewApp() (*App, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	redisClient, err := initRedisClient(cfg)
	if err != nil {
		return nil, err
	}

	healthCheck := initHealthService(cfg, redisClient)
	link := initLinkService(redisClient)

	router := api.NewRouter(cfg.AppPort, healthCheck, link)

	return &App{
		router: router,
	}, nil
}

func initRedisClient(cfg *config.Config) (*redis.RedisClient, error) {
	return redis.NewRedisClient(
		redis.RedisConfig{
			Host:     cfg.Host,
			Port:     cfg.Port,
			Password: cfg.Password,
			Database: cfg.Database,
		},
	)
}

func initHealthService(cfg *config.Config, redisClient *redis.RedisClient) healthHandler.HealthCheck {
	svc := healthService.NewHealthCheckService(
		cfg.ServiceName,
		cfg.InstanceID,
		redisClient,
	)
	return healthHandler.NewHealthCheckHandler(svc)
}

func initLinkService(redisClient *redis.RedisClient) linkHandler.Link {
	linkRepository := repository.NewLinkRepository(redisClient)
	svc := linkService.NewLinkService(linkRepository)
	return linkHandler.NewLinkHandler(svc)
}

// Run starts the application by running the HTTP server.
func (a *App) Run() error {
	return a.router.Run()
}
