package cache

import (
	"context"

	bookmarkDTO "github.com/huypham67/bookmark-service/internal/dto/bookmark"
	"github.com/huypham67/bookmark-service/internal/model"
	"github.com/huypham67/bookmark-service/internal/service/bookmark"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/rs/zerolog/log"
)

// Create creates a new bookmark for the user and invalidates the relevant cache entries.
func (s *bookmarkCacheService) Create(ctx context.Context, userID string, req bookmarkDTO.CreateBookmarkRequest) (*model.Bookmark, error) {
	segment := newrelic.FromContext(ctx).StartSegment("service.bookmark.cache.Create")
	defer segment.End()

	hashKey := buildUserCacheKey(userID)
	if err := s.cacheRepo.DeleteCacheByHashKey(ctx, hashKey); err != nil {
		log.Error().
			Err(err).
			Str("user_id", userID).
			Msg("failed to invalidate user cache; aborting bookmark creation")
		return nil, bookmark.ErrInternalServerError
	}

	bm, err := s.bookmarkService.Create(ctx, userID, req)
	if err != nil {
		return nil, err
	}

	return bm, nil
}
