package bookmark

import (
	"context"

	"github.com/huypham67/bookmark-common/pkg/dbutils"
	"github.com/huypham67/bookmark-service/internal/model"
)

// GetURLByCode retrieves the original URL associated with a given bookmark code.
func (r *repository) GetURLByCode(ctx context.Context, code string) (string, error) {
	var bookmark model.Bookmark
	if err := r.db.WithContext(ctx).
		Select("url").
		Where("code = ?", code).
		First(&bookmark).Error; err != nil {
		return "", dbutils.ClassifyError(err)
	}
	return bookmark.URL, nil
}

// GetPaginatedByUserID fetches a paginated list of bookmarks for a specific user.
func (r *repository) GetPaginatedByUserID(ctx context.Context, userID string, offset, limit int64, sort string) ([]*model.Bookmark, error) {
	var bookmarks []*model.Bookmark

	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order(sort + " DESC").
		Offset(int(offset)).
		Limit(int(limit)).
		Find(&bookmarks).Error; err != nil {
		return nil, dbutils.ClassifyError(err)
	}

	return bookmarks, nil
}

// CountByUserID counts the total number of bookmarks for a specific user.
func (r *repository) CountByUserID(ctx context.Context, userID string) (int64, error) {
	var count int64

	if err := r.db.WithContext(ctx).
		Model(&model.Bookmark{}).
		Where("user_id = ?", userID).
		Count(&count).Error; err != nil {
		return 0, dbutils.ClassifyError(err)
	}

	return count, nil
}
