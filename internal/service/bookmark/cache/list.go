package cache

import (
	"context"
	"encoding/json"
	"errors"

	bookmarkDTO "github.com/huypham67/bookmark-service/internal/dto/bookmark"
	"github.com/huypham67/bookmark-service/internal/model"
	repocache "github.com/huypham67/bookmark-service/internal/repository/cache"
	"github.com/huypham67/bookmark-service/internal/service/bookmark"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/rs/zerolog/log"
)

const (
	metricCacheHit  = "Custom/Cache/Hit"
	metricCacheMiss = "Custom/Cache/Miss"
)

type cacheData struct {
	Bookmarks  []*model.Bookmark          `json:"bookmarks"`
	Pagination *bookmark.PaginationResult `json:"pagination"`
}

// List retrieves a paginated list of bookmarks with caching.
// Attempts to retrieve from cache first; on miss, calls the bookmark service and caches the result.
func (s *bookmarkCacheService) List(ctx context.Context, userID string, req *bookmarkDTO.ListBookmarksRequest) ([]*model.Bookmark, *bookmark.PaginationResult, error) {
	txn := newrelic.FromContext(ctx)
	defer txn.StartSegment("service.bookmark.cache.List").End()

	hashKey := buildUserCacheKey(userID)
	fieldKey := buildQueryCacheKey(req.Page, req.Limit, req.Sort)

	cacheDataRaw, err := s.cacheRepo.GetCache(ctx, hashKey, fieldKey)
	if err != nil && !errors.Is(err, repocache.ErrCacheMiss) {
		log.Warn().
			Err(err).
			Str("user_id", userID).
			Int64("page", req.Page).
			Msg("cache retrieval error")
	}

	if cacheDataRaw != nil {
		var cacheResult cacheData
		if err := json.Unmarshal(cacheDataRaw, &cacheResult); err != nil {
			log.Warn().
				Err(err).
				Str("user_id", userID).
				Int64("page", req.Page).
				Msg("cache deserialization error")
			if delErr := s.cacheRepo.DeleteCacheByFieldKey(ctx, hashKey, fieldKey); delErr != nil {
				log.Warn().
					Err(delErr).
					Str("user_id", userID).
					Msg("failed to delete corrupted cache entry")
			}
		} else {
			txn.AddAttribute("cache.result", "hit")
			txn.Application().RecordCustomMetric(metricCacheHit, 1)
			return cacheResult.Bookmarks, cacheResult.Pagination, nil
		}
	}

	txn.AddAttribute("cache.result", "miss")
	txn.Application().RecordCustomMetric(metricCacheMiss, 1)
	// Cache miss: call the bookmark service
	bookmarks, pagination, err := s.bookmarkService.List(ctx, userID, req)
	if err != nil {
		return nil, nil, err
	}

	// Cache the result
	data := cacheData{
		Bookmarks:  bookmarks,
		Pagination: pagination,
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		log.Error().
			Err(err).
			Str("user_id", userID).
			Msg("failed to serialize bookmarks for cache")
	} else {
		if err := s.cacheRepo.SetCache(ctx, hashKey, fieldKey, dataBytes, cacheTTL); err != nil {
			log.Warn().
				Err(err).
				Str("user_id", userID).
				Msg("failed to cache bookmarks")
		}
	}

	log.Info().
		Str("user_id", userID).
		Int64("page", req.Page).
		Str("source", "service").
		Msg("bookmarks retrieved from service")

	return bookmarks, pagination, nil
}
