package integration

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/huypham67/bookmark-service/internal/api"
	"github.com/huypham67/bookmark-service/internal/handler"
	"github.com/huypham67/bookmark-service/internal/repository"
	"github.com/huypham67/bookmark-service/internal/repository/ping"
	"github.com/huypham67/bookmark-service/internal/repository/testutil"
	"github.com/huypham67/bookmark-service/internal/service"
	"github.com/huypham67/bookmark-service/pkg/jwtutils"
	"github.com/huypham67/bookmark-service/pkg/redis"
	"github.com/huypham67/bookmark-service/pkg/security"
	"github.com/huypham67/bookmark-service/pkg/utils"
	"github.com/stretchr/testify/require"
)

// TestApp represents the test application with its dependencies.
type TestApp struct {
	Router    *api.Router
	MockRedis *redis.MockRedis
}

func setupHealthCheckTestApp(t *testing.T, serviceName string, instanceID string) *TestApp {
	t.Helper()

	mockRedis := redis.NewMockRedis(t)

	pinger := ping.NewPinger(mockRedis.Client)

	healthService := service.NewHealthCheckService(serviceName, instanceID, pinger)

	healthHandler := handler.NewHealthCheckHandler(healthService)

	router := api.NewRouter()

	api.RegisterHealthRoutes(
		router.GroupAPI(),
		healthHandler,
	)

	return &TestApp{
		Router:    router,
		MockRedis: mockRedis,
	}
}

func setupLinkTestApp(t *testing.T) *TestApp {
	t.Helper()

	mockRedis := redis.NewMockRedis(t)

	linkRepository := repository.NewLinkRepository(mockRedis.Client)

	linkService := service.NewLinkService(
		linkRepository,
		utils.NewCodeGenerator(),
	)

	linkHandler := handler.NewLinkHandler(
		linkService,
	)

	router := api.NewRouter()

	api.RegisterLinkRoutes(
		router.GroupV1(),
		linkHandler,
	)

	return &TestApp{
		Router:    router,
		MockRedis: mockRedis,
	}
}

func createTestTokenGenerator(t *testing.T) jwtutils.TokenGenerator {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	generator, err := jwtutils.NewTokenGenerator(
		privateKey,
		"test-issuer",
		"test-audience",
		time.Hour,
	)
	require.NoError(t, err)

	return generator
}

func setupAuthTestApp(t *testing.T) *TestApp {
	t.Helper()

	mockDB := testutil.SetupUserTestDatabase(t)

	userRepository := repository.NewUserRepository(mockDB)

	passwordHasher := security.NewBcryptPasswordHasher()

	tokenGenerator := createTestTokenGenerator(t)

	authService := service.NewAuthService(userRepository, passwordHasher, tokenGenerator)

	authHandler := handler.NewAuthHandler(authService)

	router := api.NewRouter()

	api.RegisterAuthRoutes(
		router.GroupV1(),
		authHandler,
	)

	return &TestApp{
		Router: router,
	}
}
