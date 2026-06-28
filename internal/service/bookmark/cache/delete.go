package cache

import (
	"context"

	"github.com/huypham67/bookmark-service/internal/service/bookmark"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/rs/zerolog/log"
)

// Delete removes a bookmark and invalidates all related cache entries for the user.
func (s *bookmarkCacheService) Delete(ctx context.Context, userID, bookmarkID string) error {
	segment := newrelic.FromContext(ctx).StartSegment("service.bookmark.cache.Delete")
	defer segment.End()

	hashKey := buildUserCacheKey(userID)
	if err := s.cacheRepo.DeleteCacheByHashKey(ctx, hashKey); err != nil {
		log.Error().
			Err(err).
			Str("user_id", userID).
			Msg("failed to invalidate user cache; aborting bookmark deletion")
		return bookmark.ErrInternalServerError
	}

	if err := s.bookmarkService.Delete(ctx, userID, bookmarkID); err != nil {
		return err
	}

	return nil
}
