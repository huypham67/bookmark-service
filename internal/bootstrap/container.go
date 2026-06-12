package bootstrap

import (
	"github.com/gin-gonic/gin"
	"github.com/huypham67/bookmark-common/middleware"
	ratelimitprovider "github.com/huypham67/bookmark-common/pkg/ratelimit/provider"
	pkgRedis "github.com/huypham67/bookmark-common/pkg/redis"
	"github.com/huypham67/bookmark-common/pkg/sqldb"
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
	bookmarkSvc "github.com/huypham67/bookmark-service/internal/service/bookmark"
	bookmarkCacheSvc "github.com/huypham67/bookmark-service/internal/service/bookmark/cache"
	healthSvc "github.com/huypham67/bookmark-service/internal/service/health"
	linkSvc "github.com/huypham67/bookmark-service/internal/service/link"
)

// Container holds the application's dependencies and initialized services.
// It serves as the single source of truth for all infrastructure and business logic components.
type Container struct {
	// Infrastructure
	Config    *Config
	DB        *gorm.DB
	Redis     *redis.Client
	CacheRepo cacheRepo.Repository

	// Handlers
	HealthHandler   healthHandler.Handler
	LinkHandler     linkHandler.Handler
	BookmarkHandler bookmarkHandler.Handler

	// Middleware
	JWTMiddleware       gin.HandlerFunc
	RateLimitMiddleware gin.HandlerFunc
}

// NewContainer initializes the application container by loading configuration,
// setting up infrastructure clients, and initializing all handlers with their dependencies.
func NewContainer() (*Container, error) {
	cfg, err := NewConfig()
	if err != nil {
		log.Error().Err(err).Msg("failed to load config")
		return nil, err
	}

	db, err := sqldb.NewClient("")
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

	tokenValidator, err := jwtprovider.NewValidator("")
	if err != nil {
		log.Error().Err(err).Msg("failed to initialize jwt validator")
		return nil, err
	}

	jwtMiddleware := middleware.JWTAuth(tokenValidator)

	rateLimiter, err := ratelimitprovider.New(rdb, "")
	if err != nil {
		log.Error().Err(err).Msg("failed to initialize rate limiter")
		return nil, err
	}
	rateLimitMiddleware := middleware.RateLimit(rateLimiter)

	// Initialize shared infrastructure
	cacheRepository := cacheRepo.NewRedis(rdb)

	healthHandlerInstance := initHealthHandler(cfg, db, rdb)
	linkHandlerInstance := initLinkHandler(rdb, db)
	bookmarkHandlerInstance := initBookmarkHandler(db, cacheRepository)

	return &Container{
		Config:              cfg,
		DB:                  db,
		Redis:               rdb,
		CacheRepo:           cacheRepository,
		HealthHandler:       healthHandlerInstance,
		LinkHandler:         linkHandlerInstance,
		BookmarkHandler:     bookmarkHandlerInstance,
		JWTMiddleware:       jwtMiddleware,
		RateLimitMiddleware: rateLimitMiddleware,
	}, nil
}

func initHealthHandler(cfg *Config, db *gorm.DB, redisClient *redis.Client) healthHandler.Handler {
	// bookmark-service depends on both PostgreSQL (bookmarks) and Redis (links + cache),
	// so the health check pings both.
	pinger := ping.NewMulti(
		ping.NewSQLDB(db),
		ping.NewRedis(redisClient),
	)
	healthService := healthSvc.NewService(cfg.ServiceName, cfg.InstanceID, pinger)
	return healthHandler.NewHandler(healthService)
}

func initLinkHandler(redisClient *redis.Client, db *gorm.DB) linkHandler.Handler {
	linkRepository := linkRepo.NewRepository(redisClient)
	codeGenerator := utils.NewCodeGenerator()
	bookmarkResolver := bookmarkRepo.NewRepository(db)
	service := linkSvc.NewService(linkRepository, codeGenerator, bookmarkResolver)
	return linkHandler.NewHandler(service)
}

func initBookmarkHandler(db *gorm.DB, cacheRepository cacheRepo.Repository) bookmarkHandler.Handler {
	bookmarkRepository := bookmarkRepo.NewRepository(db)
	bookmarkService := bookmarkSvc.NewService(bookmarkRepository)

	cacheService := bookmarkCacheSvc.NewBookmarkService(bookmarkService, cacheRepository)

	return bookmarkHandler.NewHandler(cacheService)
}

// Close gracefully shuts down the database and Redis clients, ensuring that all resources are properly released.
func (c *Container) Close() error {
	if c.Redis != nil {
		_ = c.Redis.Close()
	}

	if c.DB != nil {
		sqlDB, err := c.DB.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	}

	return nil
}
