package cache

import (
	"fmt"
	"time"

	"github.com/huypham67/bookmark-service/internal/repository/cache"
	"github.com/huypham67/bookmark-service/internal/service/bookmark"
)

const (
	cacheNamespace = "bookmarks"
	cacheTTL       = 5 * time.Minute
)

type bookmarkCacheService struct {
	bookmarkService bookmark.Service
	cacheRepo       cache.Repository
}

// NewBookmarkService creates a new instance of the bookmark cache service with the provided dependencies.
func NewBookmarkService(bookmarkService bookmark.Service, cacheRepo cache.Repository) bookmark.Service {
	return &bookmarkCacheService{
		bookmarkService: bookmarkService,
		cacheRepo:       cacheRepo,
	}
}

func buildUserCacheKey(userID string) string {
	return fmt.Sprintf("%s:%s", cacheNamespace, userID)
}

func buildQueryCacheKey(page, limit int64, sort string) string {
	return fmt.Sprintf("page:%d:limit:%d:sort:%s", page, limit, sort)
}
