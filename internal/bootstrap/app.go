package bootstrap

import (
	"github.com/huypham67/bookmark-management/internal/api"
	"github.com/huypham67/bookmark-management/internal/config"
	"github.com/huypham67/bookmark-management/internal/handler"
	"github.com/huypham67/bookmark-management/internal/repository"
	"github.com/huypham67/bookmark-management/internal/service"
	"github.com/huypham67/bookmark-management/internal/utils"
	"github.com/huypham67/bookmark-management/pkg/logger"
	"github.com/huypham67/bookmark-management/pkg/redis"
	"github.com/rs/zerolog/log"
)

// App represents the application container holding the router.
type App struct {
	router *api.Router
}

// NewApp initializes and returns a new application instance with all dependencies configured.
func NewApp() (*App, error) {
	if err := logger.Init(""); err != nil {
		return nil, err
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Error().Err(err).Msg("Failed to load config")
		return nil, err
	}

	redisClient, err := initRedisClient()
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize Redis client")
		return nil, err
	}

	pinger := redis.NewPinger(redisClient.Client)
	healthCheckHandler := initHealthService(cfg, pinger)
	linkHandler := initLinkService(redisClient)

	router := api.NewRouter(cfg.AppPort)

	apiGroup := router.GroupV1()

	api.RegisterHealthRoutes(
		apiGroup,
		healthCheckHandler,
	)

	api.RegisterLinkRoutes(
		apiGroup,
		linkHandler,
	)

	log.Info().Msg("Application initialized successfully")

	return &App{
		router: router,
	}, nil
}

func initRedisClient() (*redis.RedisClient, error) {
	return redis.NewRedisClient("")
}

func initHealthService(cfg *config.Config, pinger redis.Pinger) handler.HealthCheck {
	svc := service.NewHealthCheckService(
		cfg.ServiceName,
		cfg.InstanceID,
		pinger,
	)
	return handler.NewHealthCheckHandler(svc)
}

func initLinkService(redisClient *redis.RedisClient) handler.Link {
	repo := repository.NewLinkRepository(redisClient)
	codeGenerator := utils.NewCodeGenerator()
	svc := service.NewLinkService(repo, codeGenerator)
	return handler.NewLinkHandler(svc)
}

// Run starts the application by running the HTTP server.
func (a *App) Run() error {
	return a.router.Run()
}
