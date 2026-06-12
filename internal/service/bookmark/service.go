package bookmark

import (
	"context"
	"errors"

	bookmarkDTO "github.com/huypham67/bookmark-service/internal/dto/bookmark"
	"github.com/huypham67/bookmark-service/internal/model"
	"github.com/huypham67/bookmark-service/internal/repository/bookmark"
)

var (
	ErrBadRequest            = errors.New("bad request")
	ErrBookmarkNotFound      = errors.New("bookmark not found")
	ErrBookmarkAlreadyExists = errors.New("bookmark code already exists")

	ErrInternalServerError = errors.New("internal server error")
)

// PaginationResult holds pagination metadata.
type PaginationResult struct {
	Page  int64
	Limit int64
	Total int64
}

// Service defines the interface for bookmark operations.
//
//go:generate mockery --name=Service --output=./mocks --outpkg=mocks --filename=mock_service.go
type Service interface {
	Create(ctx context.Context, userID string, req bookmarkDTO.CreateBookmarkRequest) (*model.Bookmark, error)
	List(ctx context.Context, userID string, req *bookmarkDTO.ListBookmarksRequest) ([]*model.Bookmark, *PaginationResult, error)
	Update(ctx context.Context, userID, bookmarkID string, req bookmarkDTO.UpdateBookmarkRequest) error
	Delete(ctx context.Context, userID, bookmarkID string) error
}

type service struct {
	bookmarkRepo bookmark.Repository
}

// NewService creates a new instance of the bookmark service with the provided dependencies.
func NewService(bookmarkRepo bookmark.Repository) Service {
	return &service{
		bookmarkRepo: bookmarkRepo,
	}
}
