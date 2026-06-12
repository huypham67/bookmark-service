package bookmark

import (
	"context"

	"github.com/huypham67/bookmark-service/internal/model"
	"gorm.io/gorm"
)

// Repository defines the interface for bookmark repository.
//
//go:generate mockery --name=Repository --output=./mocks --outpkg=mocks --filename=mock_repo.go
type Repository interface {
	Create(ctx context.Context, bookmark *model.Bookmark) error
	NextCodeInt(ctx context.Context) (int64, error)
	GetURLByCode(ctx context.Context, code string) (string, error)
	GetPaginatedByUserID(ctx context.Context, userID string, offset, limit int64, sort string) ([]*model.Bookmark, error)
	CountByUserID(ctx context.Context, userID string) (int64, error)
	Update(ctx context.Context, id, userID string, updates *model.Bookmark) (int64, error)
	Delete(ctx context.Context, id, userID string) (int64, error)
}

type repository struct {
	db *gorm.DB
}

// NewRepository creates a new bookmark repository with the given GORM database client.
func NewRepository(db *gorm.DB) Repository {
	return &repository{
		db: db,
	}
}
