package bookmark

import (
	"context"
	"errors"

	"github.com/huypham67/bookmark-common/pkg/dbutils"
	"github.com/rs/zerolog/log"
)

// Delete deletes an existing bookmark for the user.
func (s *service) Delete(ctx context.Context, userID, bookmarkID string) error {
	rowsAffected, err := s.bookmarkRepo.Delete(ctx, bookmarkID, userID)
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
				Msg("failed to delete bookmark")
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
		Msg("bookmark deleted successfully")

	return nil
}
