package bootstrap

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/huypham67/bookmark-service/docs"
	"github.com/huypham67/bookmark-service/internal/api"
	"github.com/huypham67/bookmark-service/internal/handler"
	"github.com/huypham67/bookmark-service/internal/middleware"
	"github.com/huypham67/bookmark-service/internal/model"
	"github.com/huypham67/bookmark-service/internal/repository"
	"github.com/huypham67/bookmark-service/internal/repository/ping"
	"github.com/huypham67/bookmark-service/internal/service"
	"github.com/huypham67/bookmark-service/pkg/jwtutils"
	"github.com/huypham67/bookmark-service/pkg/logger"
	pkgRedis "github.com/huypham67/bookmark-service/pkg/redis"
	"github.com/huypham67/bookmark-service/pkg/security"
	"github.com/huypham67/bookmark-service/pkg/sqldb"
	"github.com/huypham67/bookmark-service/pkg/utils"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// App represents application runtime container.
type App struct {
	config      *Config
	router      *api.Router
	redisClient *redis.Client
	dbClient    *gorm.DB
}

// NewApp initializes application dependencies.
func NewApp() (*App, error) {
	if err := logger.NewLoggerClient(""); err != nil {
		return nil, err
	}

	cfg, err := LoadConfig()

	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to load config")

		return nil, err
	}

	setupSwaggerConfig(cfg)

	redisClient, err := initRedisClient()

	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to initialize redis client")

		return nil, err
	}

	dbClient, err := initPostgresClient()

	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to initialize postgres client")

		return nil, err
	}

	// Run database migrations
	if err := runMigrations(dbClient); err != nil {
		log.Error().
			Err(err).
			Msg("failed to run database migrations")

		return nil, err
	}

	router := api.NewRouter()

	registerRoutes(router, cfg, redisClient, dbClient)

	log.Info().
		Msg("application initialized successfully")

	return &App{
		config:      cfg,
		router:      router,
		redisClient: redisClient,
	}, nil
}

func registerRoutes(router *api.Router, cfg *Config, redisClient *redis.Client, dbClient *gorm.DB) {
	apiGroup := router.GroupAPI()
	apiV1Group := router.GroupV1()

	healthHandler := initHealthHandler(cfg, redisClient)
	linkHandler := initLinkHandler(redisClient)
	userHandler := initUserHandler(dbClient)

	// Initialize JWT middleware for protected routes
	jwtMiddleware := initJWTMiddleware()

	api.RegisterHealthRoutes(apiGroup, healthHandler)
	api.RegisterLinkRoutes(apiV1Group, linkHandler)
	api.RegisterUserRoutes(apiV1Group, userHandler, jwtMiddleware)
}

func initRedisClient() (*redis.Client, error) {
	return pkgRedis.NewRedisClient("")
}

func initPostgresClient() (*gorm.DB, error) {
	return sqldb.NewDBClient("")
}

func runMigrations(db *gorm.DB) error {
	return db.AutoMigrate(&model.User{})
}

func initJWTMiddleware() gin.HandlerFunc {
	// Load public key for JWT token validation
	publicKey, err := jwtutils.LoadRSAPublicKeyFromFile("keys/public.pem")
	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to load public key for JWT")
		return func(c *gin.Context) {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
			})
			c.Abort()
		}
	}

	tokenValidator, err := jwtutils.NewTokenValidator(publicKey, "bookmark-service", "bookmark-service")
	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to create token validator")
		return func(c *gin.Context) {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
			})
			c.Abort()
		}
	}

	return middleware.JWTAuth(tokenValidator)
}

func initUserHandler(db *gorm.DB) handler.User {
	userRepository := repository.NewUserRepository(db)
	passwordHasher := security.NewBcryptPasswordHasher()

	// Load private key for JWT token generation
	privateKey, err := jwtutils.LoadRSAPrivateKeyFromFile("keys/private.pem")
	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to load private key for JWT")
		return nil
	}

	tokenGenerator, err := jwtutils.NewTokenGenerator(privateKey, "bookmark-service", "bookmark-service", time.Hour*1)
	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to create token generator")
		return nil
	}

	userService := service.NewUserService(userRepository, passwordHasher, tokenGenerator)
	return handler.NewUserHandler(userService)
}

func initHealthHandler(cfg *Config, redisClient *redis.Client) handler.HealthCheck {
	pinger := ping.NewPinger(redisClient)

	healthService := service.NewHealthCheckService(cfg.ServiceName, cfg.InstanceID, pinger)

	return handler.NewHealthCheckHandler(healthService)
}

func initLinkHandler(redisClient *redis.Client) handler.Link {
	linkRepository := repository.NewLinkRepository(redisClient)

	codeGenerator := utils.NewCodeGenerator()

	linkService := service.NewLinkService(linkRepository, codeGenerator)

	return handler.NewLinkHandler(linkService)
}

func setupSwaggerConfig(cfg *Config) {
	docs.SwaggerInfo.Host = ""
	docs.SwaggerInfo.Schemes = getSwaggerSchemes(cfg)
	docs.SwaggerInfo.BasePath = cfg.HostName
}

func getSwaggerSchemes(cfg *Config) []string {
	if cfg.SwaggerSchemes != "" {
		return parseSchemes(cfg.SwaggerSchemes)
	}

	if cfg.Environment == "production" {
		return []string{"https"}
	}

	return []string{"http"}
}

func parseSchemes(schemesStr string) []string {
	schemes := make([]string, 0)
	for _, scheme := range strings.Split(schemesStr, ",") {
		if trimmed := strings.TrimSpace(scheme); trimmed != "" {
			schemes = append(schemes, trimmed)
		}
	}
	return schemes
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
		_ = a.redisClient.Close()
	}

	if a.dbClient != nil {
		sqlDB, err := a.dbClient.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	}

	return nil
}
