package bookmark

import (
	"context"
	"errors"

	"github.com/huypham67/bookmark-common/pkg/dbutils"
	bookmarkDTO "github.com/huypham67/bookmark-service/internal/dto/bookmark"
	"github.com/huypham67/bookmark-service/internal/model"
	"github.com/rs/zerolog/log"
)

// Create creates a new bookmark for the user.
func (s *service) Create(ctx context.Context, userID string, req bookmarkDTO.CreateBookmarkRequest) (*model.Bookmark, error) {

	bm := &model.Bookmark{
		Description: req.Description,
		URL:         req.URL,
		UserID:      userID,
	}

	if err := s.bookmarkRepo.Create(ctx, bm); err != nil {
		switch {
		case errors.Is(err, dbutils.ErrDuplicationType):
			log.Warn().
				Str("user_id", userID).
				Str("url", req.URL).
				Msg("bookmark code already exists")
			return nil, ErrBookmarkAlreadyExists
		case errors.Is(err, dbutils.ErrForeignKeyViolationType):
			log.Warn().
				Str("user_id", userID).
				Msg("user not found")
			return nil, ErrBadRequest
		default:
			log.Error().
				Err(err).
				Str("user_id", userID).
				Str("url", req.URL).
				Msg("failed to create bookmark")
			return nil, ErrInternalServerError
		}
	}

	log.Info().
		Str("bookmark_id", bm.ID).
		Str("user_id", userID).
		Str("code", bm.Code).
		Msg("bookmark created successfully")

	return bm, nil
}
