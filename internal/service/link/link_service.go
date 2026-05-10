package link

import (
	"github.com/huypham67/bookmark-management/internal/dto/request"
	"github.com/huypham67/bookmark-management/internal/repository"
	"github.com/huypham67/bookmark-management/internal/utils"
)

// Link defines the contract for link services.
type Link interface {
	ShortenURL(request request.ShortenURLRequest) (string, error)
}

type linkService struct {
	linkRepo repository.Link
}

// NewLinkService creates a new link service with the given repository.
func NewLinkService(linkRepo repository.Link) Link {
	return &linkService{
		linkRepo: linkRepo,
	}
}

const shortCodeLength = 7

// ShortenURL generates a random short code for the given URL and saves it to the repository.
// If the generated code already exists, it recursively generates a new one.
func (link *linkService) ShortenURL(request request.ShortenURLRequest) (string, error) {
	code, err := utils.GenerateRandomCode(shortCodeLength)

	if err != nil {
		return "", err
	}

	exists, err := link.linkRepo.CheckExists(code)

	if err != nil {
		return "", err
	}

	if exists {
		return link.ShortenURL(request)
	}
	err = link.linkRepo.SaveLink(code, request.Url, request.Exp)

	if err != nil {
		return "", err
	}
	return code, nil
}
