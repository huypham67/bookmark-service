package bookmark

import (
	"context"
	"errors"

	"github.com/huypham67/bookmark-common/pkg/dbutils"
	bookmarkDTO "github.com/huypham67/bookmark-service/internal/dto/bookmark"
	"github.com/huypham67/bookmark-service/internal/model"
	"github.com/rs/zerolog/log"
)

// Update updates an existing bookmark for the user.
func (s *service) Update(ctx context.Context, userID, bookmarkID string, req bookmarkDTO.UpdateBookmarkRequest) error {
	updates := &model.Bookmark{}
	if req.Description != "" {
		updates.Description = req.Description
	}
	if req.URL != "" {
		updates.URL = req.URL
	}

	rowsAffected, err := s.bookmarkRepo.Update(ctx, bookmarkID, userID, updates)
	if err != nil {
		switch {
		case errors.Is(err, dbutils.ErrForeignKeyViolationType):
			log.Warn().
				Str("user_id", userID).
				Str("bookmark_id", bookmarkID).
				Msg("user not found")
			return ErrBadRequest
		default:
			log.Error().
				Err(err).
				Str("user_id", userID).
				Str("bookmark_id", bookmarkID).
				Msg("failed to update bookmark")
			return ErrInternalServerError
		}
	}

	if rowsAffected == 0 {
		log.Warn().
			Str("user_id", userID).
			Str("bookmark_id", bookmarkID).
			Msg("bookmark not found")
		return ErrBookmarkNotFound
	}

	log.Info().
		Str("bookmark_id", bookmarkID).
		Str("user_id", userID).
		Msg("bookmark updated successfully")

	return nil
}
