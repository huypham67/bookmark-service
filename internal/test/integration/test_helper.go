package integration

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/huypham67/bookmark-common/middleware"
	"github.com/stretchr/testify/require"

	"github.com/huypham67/bookmark-common/pkg/jwt"
	pkgRedis "github.com/huypham67/bookmark-common/pkg/redis"
	"github.com/huypham67/bookmark-common/pkg/sqldb"
	"github.com/huypham67/bookmark-common/pkg/utils"
	"github.com/huypham67/bookmark-service/internal/api"
	bookmarkHandler "github.com/huypham67/bookmark-service/internal/handler/bookmark"
	healthHandler "github.com/huypham67/bookmark-service/internal/handler/health"
	linkHandler "github.com/huypham67/bookmark-service/internal/handler/link"
	"github.com/huypham67/bookmark-service/internal/model"
	bookmarkRepo "github.com/huypham67/bookmark-service/internal/repository/bookmark"
	cacheRepo "github.com/huypham67/bookmark-service/internal/repository/cache"
	linkRepo "github.com/huypham67/bookmark-service/internal/repository/link"
	"github.com/huypham67/bookmark-service/internal/repository/ping"
	bookmarkSvc "github.com/huypham67/bookmark-service/internal/service/bookmark"
	bookmarkCacheSvc "github.com/huypham67/bookmark-service/internal/service/bookmark/cache"
	healthSvc "github.com/huypham67/bookmark-service/internal/service/health"
	linkSvc "github.com/huypham67/bookmark-service/internal/service/link"
	"github.com/huypham67/bookmark-service/internal/test/fixtures"
	"gorm.io/gorm"
)

const (
	testIssuer   = "test-issuer"
	testAudience = "test-audience"
)

// TestApp represents the test application with its dependencies.
type TestApp struct {
	Router    *api.Router
	MockRedis *pkgRedis.Mock
	MockDB    *gorm.DB
}

type AuthenticatedTestApp struct {
	*TestApp
	TokenGenerator jwt.TokenGenerator
}

func createTestJWT(t *testing.T) (
	jwt.TokenGenerator,
	jwt.TokenValidator,
) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(
		rand.Reader,
		2048,
	)
	require.NoError(t, err)

	tokenGenerator, err := jwt.NewTokenGenerator(
		privateKey,
		testIssuer,
		testAudience,
		time.Hour,
	)
	require.NoError(t, err)

	tokenValidator, err := jwt.NewTokenValidator(
		&privateKey.PublicKey,
		testIssuer,
		testAudience,
	)
	require.NoError(t, err)

	return tokenGenerator, tokenValidator
}

func setupHealthCheckTestApp(t *testing.T, serviceName string, instanceID string) *TestApp {
	t.Helper()

	mockRedis := pkgRedis.NewMock(t)
	mockDB := sqldb.NewMock(t)

	// bookmark-service health check pings both the database and Redis.
	pinger := ping.NewMulti(
		ping.NewSQLDB(mockDB),
		ping.NewRedis(mockRedis.Client),
	)

	healthService := healthSvc.NewService(serviceName, instanceID, pinger)

	healthHandlerInstance := healthHandler.NewHandler(healthService)

	router := api.NewRouter()

	api.RegisterHealthRoutes(
		router.GroupAPI(),
		healthHandlerInstance,
	)

	return &TestApp{
		Router:    router,
		MockRedis: mockRedis,
		MockDB:    mockDB,
	}
}

func setupLinkTestApp(t *testing.T) *TestApp {
	t.Helper()

	mockRedis := pkgRedis.NewMock(t)

	linkRepository := linkRepo.NewRepository(mockRedis.Client)

	mockDB := fixtures.NewTestDB(t, &fixtures.BookmarkTestDB{})
	bookmarkResolver := bookmarkRepo.NewRepository(mockDB)

	linkService := linkSvc.NewService(
		linkRepository,
		utils.NewCodeGenerator(),
		bookmarkResolver,
	)

	linkHandlerInstance := linkHandler.NewHandler(
		linkService,
	)

	router := api.NewRouter()

	api.RegisterLinkRoutes(
		router.GroupV1(),
		linkHandlerInstance,
	)

	return &TestApp{
		Router:    router,
		MockRedis: mockRedis,
	}
}

func setupBookmarkTestApp(t *testing.T) *AuthenticatedTestApp {
	t.Helper()

	mockDB := fixtures.NewTestDB(t, &fixtures.BookmarkTestDB{})
	mockRedis := pkgRedis.NewMock(t)

	bookmarkRepository := bookmarkRepo.NewRepository(mockDB)
	bookmarkService := bookmarkSvc.NewService(bookmarkRepository)

	cacheRepository := cacheRepo.NewRedis(mockRedis.Client)
	cacheService := bookmarkCacheSvc.NewBookmarkService(bookmarkService, cacheRepository)

	bookmarkHandlerInstance := bookmarkHandler.NewHandler(cacheService)

	tokenGenerator, tokenValidator := createTestJWT(t)

	router := api.NewRouter()

	jwtMiddleware := middleware.JWTAuth(tokenValidator)

	api.RegisterBookmarkRoutes(router.GroupV1(), bookmarkHandlerInstance, jwtMiddleware)

	return &AuthenticatedTestApp{
		TestApp: &TestApp{
			Router:    router,
			MockRedis: mockRedis,
		},
		TokenGenerator: tokenGenerator,
	}
}

const cacheSeedUserID = "user-uuid-1"

func cacheSeededBookmarks() []*model.Bookmark {
	return []*model.Bookmark{
		{BaseModel: model.BaseModel{ID: "cache-bm-1"}, Description: "Cached Bookmark 1", URL: "https://cache.example.com/1", Code: "zCACHE1", UserID: cacheSeedUserID},
		{BaseModel: model.BaseModel{ID: "cache-bm-2"}, Description: "Cached Bookmark 2", URL: "https://cache.example.com/2", Code: "zCACHE2", UserID: cacheSeedUserID},
		{BaseModel: model.BaseModel{ID: "cache-bm-3"}, Description: "Cached Bookmark 3", URL: "https://cache.example.com/3", Code: "zCACHE3", UserID: cacheSeedUserID},
	}
}

func bookmarkCacheHashKey(userID string) string {
	return fmt.Sprintf("bookmarks:%s", userID)
}

func bookmarkCacheFieldKey(page, limit int64, sort string) string {
	return fmt.Sprintf("page:%d:limit:%d:sort:%s", page, limit, sort)
}

func seedBookmarkListCache(t *testing.T, app *AuthenticatedTestApp, page, limit int64, sort string) {
	t.Helper()

	bookmarks := cacheSeededBookmarks()
	payload := struct {
		Bookmarks  []*model.Bookmark             `json:"bookmarks"`
		Pagination *bookmarkSvc.PaginationResult `json:"pagination"`
	}{
		Bookmarks:  bookmarks,
		Pagination: &bookmarkSvc.PaginationResult{Page: page, Limit: limit, Total: int64(len(bookmarks))},
	}

	data, err := json.Marshal(payload)
	require.NoError(t, err)

	err = app.MockRedis.Client.HSet(
		context.Background(),
		bookmarkCacheHashKey(cacheSeedUserID),
		bookmarkCacheFieldKey(page, limit, sort),
		data,
	).Err()
	require.NoError(t, err)
}
