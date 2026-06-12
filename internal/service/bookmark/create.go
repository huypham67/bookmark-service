package bookmark

import (
	"context"
	"errors"

	"github.com/huypham67/bookmark-common/pkg/dbutils"
	"github.com/huypham67/bookmark-common/pkg/shortcode"
	bookmarkDTO "github.com/huypham67/bookmark-service/internal/dto/bookmark"
	"github.com/huypham67/bookmark-service/internal/model"
	"github.com/rs/zerolog/log"
)

// Create creates a new bookmark for the user.
//
// The code is derived from an auto-increment sequence value (code_int) rather
// than a random string, so it is guaranteed unique: code_int never repeats, so
// its base62 code never collides.
func (s *service) Create(ctx context.Context, userID string, req bookmarkDTO.CreateBookmarkRequest) (*model.Bookmark, error) {
	codeInt, err := s.bookmarkRepo.NextCodeInt(ctx)
	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to reserve bookmark code_int")
		return nil, ErrInternalServerError
	}

	code, err := shortcode.EncodeSQLCode(uint64(codeInt))
	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to encode bookmark code")
		return nil, ErrInternalServerError
	}

	bm := &model.Bookmark{
		Description: req.Description,
		URL:         req.URL,
		Code:        code,
		CodeInt:     int(codeInt),
		UserID:      userID,
	}

	if err := s.bookmarkRepo.Create(ctx, bm); err != nil {
		switch {
		case errors.Is(err, dbutils.ErrDuplicationType):
			log.Warn().
				Str("user_id", userID).
				Str("code", code).
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
