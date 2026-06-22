package bootstrap

import (
	"github.com/gin-gonic/gin"
	"github.com/huypham67/bookmark-common/middleware"
	ratelimitprovider "github.com/huypham67/bookmark-common/pkg/ratelimit/provider"
	pkgRedis "github.com/huypham67/bookmark-common/pkg/redis"
	"github.com/huypham67/bookmark-common/pkg/sqldb"
	"github.com/huypham67/bookmark-common/pkg/tracing"
	"github.com/newrelic/go-agent/v3/integrations/nrredis-v9"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	jwtprovider "github.com/huypham67/bookmark-common/pkg/jwt/provider"
	"github.com/huypham67/bookmark-common/pkg/utils"
	bookmarkHandler "github.com/huypham67/bookmark-service/internal/handler/bookmark"
	healthHandler "github.com/huypham67/bookmark-service/internal/handler/health"
	linkHandler "github.com/huypham67/bookmark-service/internal/handler/link"
	bookmarkRepo "github.com/huypham67/bookmark-service/internal/repository/bookmark"
	cacheRepo "github.com/huypham67/bookmark-service/internal/repository/cache"
	linkRepo "github.com/huypham67/bookmark-service/internal/repository/link"
	"github.com/huypham67/bookmark-service/internal/repository/ping"
	queueRepo "github.com/huypham67/bookmark-service/internal/repository/queue"
	bookmarkSvc "github.com/huypham67/bookmark-service/internal/service/bookmark"
	bookmarkCacheSvc "github.com/huypham67/bookmark-service/internal/service/bookmark/cache"
	healthSvc "github.com/huypham67/bookmark-service/internal/service/health"
	linkSvc "github.com/huypham67/bookmark-service/internal/service/link"
)

// Container holds the application's dependencies and initialized services.
type Container struct {
	Config         *Config
	DB             *gorm.DB
	Redis          *redis.Client
	NRApp          *newrelic.Application
	CacheRepo      cacheRepo.Repository
	QueuePublisher queueRepo.Publisher

	HealthHandler   healthHandler.Handler
	LinkHandler     linkHandler.Handler
	BookmarkHandler bookmarkHandler.Handler

	JWTMiddleware       gin.HandlerFunc
	RateLimitMiddleware gin.HandlerFunc
}

// NewContainer initializes the application container.
func NewContainer() (*Container, error) {
	cfg, err := NewConfig()
	if err != nil {
		log.Error().Err(err).Msg("failed to load config")
		return nil, err
	}

	nrApp, err := tracing.NewApplication("")
	if err != nil {
		log.Error().Err(err).Msg("failed to initialize New Relic")
		return nil, err
	}

	db, err := sqldb.NewInstrumentedClient("")
	if err != nil {
		log.Error().Err(err).Msg("failed to initialize postgres client")
		return nil, err
	}

	db, err = sqldb.RunMigration(db, "migrations")
	if err != nil {
		log.Error().Err(err).Msg("failed to run database migrations")
		return nil, err
	}

	rdb, err := pkgRedis.NewClient("")
	if err != nil {
		log.Error().Err(err).Msg("failed to initialize redis client")
		return nil, err
	}
	rdb.AddHook(nrredis.NewHook(rdb.Options()))

	tokenValidator, err := jwtprovider.NewValidator("")
	if err != nil {
		log.Error().Err(err).Msg("failed to initialize jwt validator")
		return nil, err
	}

	rateLimiter, err := ratelimitprovider.New(rdb, "")
	if err != nil {
		log.Error().Err(err).Msg("failed to initialize rate limiter")
		return nil, err
	}

	cacheRepository := cacheRepo.NewRedis(rdb)
	queuePublisher := queueRepo.NewRedisPublisher(rdb)

	return &Container{
		Config:              cfg,
		DB:                  db,
		Redis:               rdb,
		NRApp:               nrApp,
		CacheRepo:           cacheRepository,
		QueuePublisher:      queuePublisher,
		HealthHandler:       initHealthHandler(cfg, db, rdb),
		LinkHandler:         initLinkHandler(rdb, db),
		BookmarkHandler:     initBookmarkHandler(db, cacheRepository, queuePublisher),
		JWTMiddleware:       middleware.JWTAuth(tokenValidator),
		RateLimitMiddleware: middleware.RateLimit(rateLimiter),
	}, nil
}

func initHealthHandler(cfg *Config, db *gorm.DB, redisClient *redis.Client) healthHandler.Handler {
	pinger := ping.NewMulti(ping.NewSQLDB(db), ping.NewRedis(redisClient))
	return healthHandler.NewHandler(healthSvc.NewService(cfg.ServiceName, cfg.InstanceID, pinger))
}

func initLinkHandler(redisClient *redis.Client, db *gorm.DB) linkHandler.Handler {
	service := linkSvc.NewService(
		linkRepo.NewRepository(redisClient),
		utils.NewCodeGenerator(),
		bookmarkRepo.NewRepository(db),
	)
	return linkHandler.NewHandler(service)
}

func initBookmarkHandler(db *gorm.DB, cacheRepository cacheRepo.Repository, queuePublisher queueRepo.Publisher) bookmarkHandler.Handler {
	bookmarkRepository := bookmarkRepo.NewRepository(db)
	cacheService := bookmarkCacheSvc.NewBookmarkService(bookmarkSvc.NewService(bookmarkRepository), cacheRepository)
	return bookmarkHandler.NewHandler(cacheService, bookmarkSvc.NewImporter(queuePublisher))
}

// Close gracefully shuts down all resources.
func (c *Container) Close() error {
	if c.Redis != nil {
		_ = c.Redis.Close()
	}
	if c.DB != nil {
		if sqlDB, err := c.DB.DB(); err == nil {
			_ = sqlDB.Close()
		}
	}
	if c.NRApp != nil {
		c.NRApp.Shutdown(0)
	}
	return nil
}
