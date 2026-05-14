package integration

import (
	"testing"

	"github.com/huypham67/bookmark-management/internal/api"
	"github.com/huypham67/bookmark-management/internal/handler"
	"github.com/huypham67/bookmark-management/internal/repository"
	"github.com/huypham67/bookmark-management/internal/service"
	"github.com/huypham67/bookmark-management/pkg/redis"
	"github.com/huypham67/bookmark-management/pkg/utils"
)

// TestApp represents the test application with its dependencies.
type TestApp struct {
	Router    *api.Router
	MockRedis *redis.MockRedis
}

func setupHealthCheckTestApp(t *testing.T, serviceName string, instanceID string) *TestApp {
	t.Helper()

	mockRedis := redis.NewMockRedis(t)

	pinger := redis.NewPinger(mockRedis.Client)

	healthService := service.NewHealthCheckService(serviceName, instanceID, pinger)

	healthHandler := handler.NewHealthCheckHandler(healthService)

	router := api.NewRouter()

	api.RegisterHealthRoutes(
		router.GroupV1(),
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

	linkRepository := repository.NewLinkRepository(
		&redis.RedisClient{
			Client: mockRedis.Client,
		},
	)

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
