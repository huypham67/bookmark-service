package bootstrap

import (
	"net/http"

	"github.com/huypham67/bookmark-service/internal/api"
	"github.com/huypham67/bookmark-service/internal/config"
	"github.com/huypham67/bookmark-service/internal/handler"
	"github.com/huypham67/bookmark-service/internal/repository"
	"github.com/huypham67/bookmark-service/internal/service"
	"github.com/huypham67/bookmark-service/pkg/logger"
	pkgRedis "github.com/huypham67/bookmark-service/pkg/redis"
	"github.com/huypham67/bookmark-service/pkg/utils"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// App represents application runtime container.
type App struct {
	config      *config.Config
	router      *api.Router
	redisClient *redis.Client
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
		config:      cfg,
		router:      router,
		redisClient: redisClient,
	}, nil
}

func registerRoutes(router *api.Router, cfg *config.Config, redisClient *redis.Client) {
	apiGroup := router.GroupAPI()
	apiV1Group := router.GroupV1()

	healthHandler := initHealthHandler(cfg, redisClient)

	linkHandler := initLinkHandler(redisClient)

	api.RegisterHealthRoutes(apiGroup, healthHandler)

	api.RegisterLinkRoutes(apiV1Group, linkHandler)
}

func initRedisClient() (*redis.Client, error) {
	return pkgRedis.NewRedisClient("")
}

func initHealthHandler(cfg *config.Config, redisClient *redis.Client) handler.HealthCheck {
	pinger := repository.NewPinger(redisClient)

	healthService := service.NewHealthCheckService(cfg.ServiceName, cfg.InstanceID, pinger)

	return handler.NewHealthCheckHandler(healthService)
}

func initLinkHandler(redisClient *redis.Client) handler.Link {
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

// Close closes all application resources.
func (a *App) Close() error {
	if a.redisClient != nil {
		return a.redisClient.Close()
	}
	return nil
}
