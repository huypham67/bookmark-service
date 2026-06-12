package bookmark

import (
	"context"

	bookmarkDTO "github.com/huypham67/bookmark-service/internal/dto/bookmark"
	"github.com/huypham67/bookmark-service/internal/model"
	"github.com/rs/zerolog/log"
)

// List retrieves a paginated list of bookmarks for the user.
func (s *service) List(ctx context.Context, userID string, req *bookmarkDTO.ListBookmarksRequest) ([]*model.Bookmark, *PaginationResult, error) {

	offset := (req.Page - 1) * req.Limit

	bookmarks, err := s.bookmarkRepo.GetPaginatedByUserID(ctx, userID, offset, req.Limit, req.Sort)
	if err != nil {
		log.Error().
			Err(err).
			Str("user_id", userID).
			Int64("page", req.Page).
			Int64("limit", req.Limit).
			Msg("failed to fetch paginated bookmarks")
		return nil, nil, ErrInternalServerError
	}

	total, err := s.bookmarkRepo.CountByUserID(ctx, userID)
	if err != nil {
		log.Error().
			Err(err).
			Str("user_id", userID).
			Msg("failed to count bookmarks")
		return nil, nil, ErrInternalServerError
	}

	log.Info().
		Str("user_id", userID).
		Int64("page", req.Page).
		Int64("limit", req.Limit).
		Int64("total", total).
		Msg("bookmarks listed successfully")

	return bookmarks, &PaginationResult{Page: req.Page, Limit: req.Limit, Total: total}, nil
}
