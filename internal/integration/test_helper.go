package integration

import (
	"testing"

	"github.com/huypham67/bookmark-service/internal/api"
	"github.com/huypham67/bookmark-service/internal/handler"
	"github.com/huypham67/bookmark-service/internal/repository"
	"github.com/huypham67/bookmark-service/internal/repository/ping"
	"github.com/huypham67/bookmark-service/internal/repository/testutil"
	"github.com/huypham67/bookmark-service/internal/service"
	"github.com/huypham67/bookmark-service/pkg/redis"
	"github.com/huypham67/bookmark-service/pkg/security"
	"github.com/huypham67/bookmark-service/pkg/utils"
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

func setupUserTestApp(t *testing.T) *TestApp {
	t.Helper()

	mockDB := testutil.SetupUserTestDatabase(t)

	userRepository := repository.NewUserRepository(mockDB)

	passwordHasher := security.NewBcryptPasswordHasher()

	userService := service.NewUserService(userRepository, passwordHasher)

	userHandler := handler.NewUserHandler(userService)

	router := api.NewRouter()

	api.RegisterUserRoutes(
		router.GroupV1(),
		userHandler,
	)

	return &TestApp{
		Router: router,
	}
}
