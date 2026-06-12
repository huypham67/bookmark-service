package cache

import (
	"context"

	bookmarkDTO "github.com/huypham67/bookmark-service/internal/dto/bookmark"
	"github.com/huypham67/bookmark-service/internal/service/bookmark"
	"github.com/rs/zerolog/log"
)

// Update updates an existing bookmark and invalidates the relevant cache entries.
func (s *bookmarkCacheService) Update(ctx context.Context, userID, bookmarkID string, req bookmarkDTO.UpdateBookmarkRequest) error {
	// Invalidate all list caches for this user before writing.
	hashKey := buildUserCacheKey(userID)
	if err := s.cacheRepo.DeleteCacheByHashKey(ctx, hashKey); err != nil {
		log.Error().
			Err(err).
			Str("user_id", userID).
			Msg("failed to invalidate user cache; aborting bookmark update")
		return bookmark.ErrInternalServerError
	}

	if err := s.bookmarkService.Update(ctx, userID, bookmarkID, req); err != nil {
		return err
	}

	return nil
}
