package service

import (
	"github.com/huypham67/bookmark-management/internal/dto/request"
	"github.com/huypham67/bookmark-management/internal/repository"
	"github.com/huypham67/bookmark-management/internal/utils"
	"github.com/huypham67/bookmark-management/pkg/logger"
)

// LinkService defines the contract for link services.
type LinkService interface {
	ShortenURL(request request.ShortenURLRequest) (string, error)
	GetOriginalURL(code string) (string, error)
}

type linkService struct {
	linkRepo      repository.Link
	codeGenerator utils.CodeGenerator
}

// NewLinkService creates a new link service with the given repository and code generator.
func NewLinkService(linkRepo repository.Link, codeGenerator utils.CodeGenerator) LinkService {
	return &linkService{
		linkRepo:      linkRepo,
		codeGenerator: codeGenerator,
	}
}

const shortCodeLength = 7

// ShortenURL generates a unique short code for the provided URL and saves the mapping to Redis with an expiration time.
func (link *linkService) ShortenURL(request request.ShortenURLRequest) (string, error) {
	code, err := link.codeGenerator.Generate(shortCodeLength)

	if err != nil {
		logger.Get().Error().
			Err(err).
			Msg("failed to generate short code")
		return "", err
	}

	exists, err := link.linkRepo.CheckExists(code)

	if err != nil {
		logger.Get().Error().
			Err(err).
			Str("code", code).
			Msg("failed to check if short code exists")
		return "", err
	}

	if exists {
		return link.ShortenURL(request)
	}
	err = link.linkRepo.SaveLink(code, request.Url, request.Exp)

	if err != nil {
		logger.Get().Error().
			Err(err).
			Str("code", code).
			Str("url", request.Url).
			Msg("failed to save short link to Redis")
		return "", err
	}
	return code, nil
}

// GetOriginalURL retrieves the original URL for a given shortened code.
func (link *linkService) GetOriginalURL(code string) (string, error) {
	url, err := link.linkRepo.GetLink(code)

	if err != nil {
		logger.Get().Error().
			Err(err).
			Str("code", code).
			Msg("failed to retrieve original URL from Redis")
		return "", err
	}

	return url, nil
}
