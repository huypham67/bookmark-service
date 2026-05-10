package bootstrap

import (
	"net/http"

	"github.com/huypham67/bookmark-management/internal/api"
	"github.com/huypham67/bookmark-management/internal/config"
	"github.com/huypham67/bookmark-management/internal/handler"
	"github.com/huypham67/bookmark-management/internal/repository"
	"github.com/huypham67/bookmark-management/internal/service"
	"github.com/huypham67/bookmark-management/pkg/logger"
	"github.com/huypham67/bookmark-management/pkg/redis"
	"github.com/huypham67/bookmark-management/pkg/utils"
	"github.com/rs/zerolog/log"
)

// App represents application runtime container.
type App struct {
	config *config.Config
	router *api.Router
}

// NewApp initializes application dependencies.
func NewApp() (*App, error) {
	if err := logger.Init(""); err != nil {
		return nil, err
	}

	cfg, err := config.LoadConfig()

	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to load config")

		return nil, err
	}

	redisClient, err := initRedisClient()

	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to initialize redis client")

		return nil, err
	}

	router := api.NewRouter()

	registerRoutes(router, cfg, redisClient)

	log.Info().
		Msg("application initialized successfully")

	return &App{
		config: cfg,
		router: router,
	}, nil
}

func registerRoutes(router *api.Router, cfg *config.Config, redisClient *redis.RedisClient) {
	apiGroup := router.GroupV1()

	healthHandler := initHealthHandler(cfg, redisClient)

	linkHandler := initLinkHandler(redisClient)

	api.RegisterHealthRoutes(apiGroup, healthHandler)

	api.RegisterLinkRoutes(apiGroup, linkHandler)
}

func initRedisClient() (*redis.RedisClient, error) {
	return redis.NewRedisClient("")
}

func initHealthHandler(cfg *config.Config, redisClient *redis.RedisClient) handler.HealthCheck {
	pinger := redis.NewPinger(redisClient.Client)

	healthService := service.NewHealthCheckService(cfg.ServiceName, cfg.InstanceID, pinger)

	return handler.NewHealthCheckHandler(healthService)
}

func initLinkHandler(redisClient *redis.RedisClient) handler.Link {
	linkRepository := repository.NewLinkRepository(redisClient)

	codeGenerator := utils.NewCodeGenerator()

	linkService := service.NewLinkService(linkRepository, codeGenerator)

	return handler.NewLinkHandler(linkService)
}

// Run starts HTTP server.
func (a *App) Run() error {
	return http.ListenAndServe(
		":"+a.config.AppPort,
		a.router,
	)
}
