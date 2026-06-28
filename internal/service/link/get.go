package link

import (
	"context"

	"github.com/huypham67/bookmark-common/pkg/dbutils"
	"github.com/huypham67/bookmark-common/pkg/shortcode"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/rs/zerolog/log"
)

// GetOriginalURL retrieves the original URL for a given shortened code.
//
// The code's routing prefix decides which store to query: Redis for shortened
// links, SQL for bookmarks. A code with no recognized prefix is treated as not
// found.
func (s *service) GetOriginalURL(ctx context.Context, code string) (string, error) {
	segment := newrelic.FromContext(ctx).StartSegment("service.link.GetOriginalURL")
	defer segment.End()

	switch shortcode.Classify(code) {
	case shortcode.StoreRedis:
		return s.getShortLinkURL(ctx, code)
	case shortcode.StoreSQL:
		return s.getBookmarkURL(ctx, code)
	default:
		log.Warn().
			Str("code", code).
			Msg("code has no recognised routing prefix")
		return "", dbutils.ErrRecordNotFoundType
	}
}

func (s *service) getShortLinkURL(ctx context.Context, code string) (string, error) {
	url, err := s.linkRepo.GetLink(ctx, code)
	if err != nil {
		log.Error().
			Err(err).
			Str("code", code).
			Msg("failed to retrieve original URL from Redis")
		return "", err
	}
	return url, nil
}

func (s *service) getBookmarkURL(ctx context.Context, code string) (string, error) {
	url, err := s.bookmarkResolver.GetURLByCode(ctx, code)
	if err != nil {
		log.Error().
			Err(err).
			Str("code", code).
			Msg("failed to retrieve original URL from bookmark store")
		return "", err
	}
	return url, nil
}
