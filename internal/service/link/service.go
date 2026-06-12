// Package link provides link shortening and retrieval services.
// It handles URL shortening with unique code generation and URL retrieval operations.
package link

import (
	"context"

	"github.com/huypham67/bookmark-common/pkg/utils"
	linkDTO "github.com/huypham67/bookmark-service/internal/dto/link"
	"github.com/huypham67/bookmark-service/internal/repository/link"
	"github.com/huypham67/bookmark-service/internal/service/link/resolver"
)

const shortCodeLength = 7

// Service defines the contract for link operations.
//
//go:generate mockery --name=Service --output=./mocks --outpkg=mocks --filename=mock_service.go
type Service interface {
	ShortenURL(ctx context.Context, request linkDTO.ShortenURLRequest) (string, error)
	GetOriginalURL(ctx context.Context, code string) (string, error)
}

type service struct {
	linkRepo         link.Repository
	codeGenerator    utils.CodeGenerator
	bookmarkResolver resolver.Bookmark
}

// NewService creates a new link service. The bookmarkResolver lets the
// redirect endpoint resolve codes that route to the SQL bookmark store.
func NewService(linkRepo link.Repository, codeGenerator utils.CodeGenerator, bookmarkResolver resolver.Bookmark) Service {
	return &service{
		linkRepo:         linkRepo,
		codeGenerator:    codeGenerator,
		bookmarkResolver: bookmarkResolver,
	}
}
