package service

import (
	"context"

	"github.com/huypham67/bookmark-service/internal/dto/request"
	"github.com/huypham67/bookmark-service/internal/repository"
	"github.com/huypham67/bookmark-service/pkg/utils"
	"github.com/rs/zerolog/log"
)

// Link defines the contract for link services.
// mockery --name=Link --dir=internal/service --output=internal/service/mocks --filename=link_service.go
type Link interface {
	ShortenURL(ctx context.Context, request request.ShortenURLRequest) (string, error)
	GetOriginalURL(ctx context.Context, code string) (string, error)
}

type linkService struct {
	linkRepo      repository.Link
	codeGenerator utils.CodeGenerator
}

// NewLinkService creates a new link service with the given repository and code generator.
func NewLinkService(linkRepo repository.Link, codeGenerator utils.CodeGenerator) Link {
	return &linkService{
		linkRepo:      linkRepo,
		codeGenerator: codeGenerator,
	}
}

const shortCodeLength = 7

// ShortenURL generates a unique short code for the provided URL and saves the mapping to Redis.
func (link *linkService) ShortenURL(ctx context.Context, request request.ShortenURLRequest) (string, error) {
	code, err := link.codeGenerator.Generate(shortCodeLength)

	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to generate short code")
		return "", err
	}

	exists, err := link.linkRepo.CheckExists(ctx, code)

	if err != nil {
		log.Error().
			Err(err).
			Str("code", code).
			Msg("failed to check if short code exists")
		return "", err
	}

	if exists {
		return link.ShortenURL(ctx, request)
	}
	err = link.linkRepo.SaveLink(ctx, code, request.Url, request.Exp)

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

// GetOriginalURL retrieves the original URL for a given shortened code.
func (link *linkService) GetOriginalURL(ctx context.Context, code string) (string, error) {
	url, err := link.linkRepo.GetLink(ctx, code)

	if err != nil {
		log.Error().
			Err(err).
			Str("code", code).
			Msg("failed to retrieve original URL from Redis")
		return "", err
	}

	return url, nil
}
