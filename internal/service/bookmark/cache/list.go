package cache

import (
	"context"
	"encoding/json"

	bookmarkDTO "github.com/huypham67/bookmark-service/internal/dto/bookmark"
	"github.com/huypham67/bookmark-service/internal/model"
	"github.com/huypham67/bookmark-service/internal/service/bookmark"
	"github.com/rs/zerolog/log"
)

type cacheData struct {
	Bookmarks  []*model.Bookmark          `json:"bookmarks"`
	Pagination *bookmark.PaginationResult `json:"pagination"`
}

// List retrieves a paginated list of bookmarks with caching.
// Attempts to retrieve from cache first; on miss, calls the bookmark service and caches the result.
func (s *bookmarkCacheService) List(ctx context.Context, userID string, req *bookmarkDTO.ListBookmarksRequest) ([]*model.Bookmark, *bookmark.PaginationResult, error) {
	hashKey := buildUserCacheKey(userID)
	fieldKey := buildQueryCacheKey(req.Page, req.Limit, req.Sort)

	// Try to get data from cache
	cacheDataRaw, err := s.cacheRepo.GetCache(ctx, hashKey, fieldKey)
	if err != nil {
		log.Warn().
			Err(err).
			Str("user_id", userID).
			Int64("page", req.Page).
			Str("source", "cache").
			Msg("cache retrieval error")
	}

	// Cache hit: deserialize and return
	if cacheDataRaw != nil {
		var cacheResult cacheData
		if err := json.Unmarshal(cacheDataRaw, &cacheResult); err != nil {
			log.Warn().
				Err(err).
				Str("user_id", userID).
				Int64("page", req.Page).
				Str("source", "cache").
				Msg("cache deserialization error")
			// Delete corrupted cache entry to prevent repeated errors
			if delErr := s.cacheRepo.DeleteCacheByFieldKey(ctx, hashKey, fieldKey); delErr != nil {
				log.Warn().
					Err(delErr).
					Str("user_id", userID).
					Msg("failed to delete corrupted cache entry")
			}
		} else {
			log.Info().
				Str("user_id", userID).
				Int64("page", req.Page).
				Str("source", "cache").
				Msg("bookmarks retrieved from cache")
			return cacheResult.Bookmarks, cacheResult.Pagination, nil
		}
	}

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
