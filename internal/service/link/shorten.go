package link

import (
	"context"

	"github.com/huypham67/bookmark-common/pkg/shortcode"
	linkDTO "github.com/huypham67/bookmark-service/internal/dto/link"
	"github.com/rs/zerolog/log"
)

// ShortenURL generates a unique short code for the provided URL and saves the mapping to Redis.
//
// The code carries a Redis routing prefix so the redirect endpoint can tell it
// apart from SQL bookmark codes.
func (s *service) ShortenURL(ctx context.Context, request linkDTO.ShortenURLRequest) (string, error) {
	payload, err := s.codeGenerator.Generate(shortCodeLength)

	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to generate short code")
		return "", err
	}

	code, err := shortcode.AddRedisPrefix(payload)
	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to add routing prefix to short code")
		return "", err
	}

	exists, err := s.linkRepo.CheckExists(ctx, code)

	if err != nil {
		log.Error().
			Err(err).
			Str("code", code).
			Msg("failed to check if short code exists")
		return "", err
	}

	if exists {
		return s.ShortenURL(ctx, request)
	}
	err = s.linkRepo.SaveLink(ctx, code, request.Url, request.Exp)

	if err != nil {
		log.Error().
			Err(err).
			Str("code", code).
			Str("url", request.Url).
			Msg("failed to save short link to Redis")
		return "", err
	}
	return code, nil
}
