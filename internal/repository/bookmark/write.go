package bookmark

import (
	"context"

	"github.com/huypham67/bookmark-common/pkg/dbutils"
	"github.com/huypham67/bookmark-common/pkg/shortcode"
	"github.com/huypham67/bookmark-service/internal/model"
	"gorm.io/gorm"
)

// Create saves a new bookmark to the database.
func (r *repository) Create(ctx context.Context, bookmark *model.Bookmark) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Create(bookmark).Error; err != nil {
			return dbutils.ClassifyError(err)
		}

		code, err := shortcode.EncodeSQLCode(uint64(bookmark.CodeInt))
		if err != nil {
			return dbutils.ClassifyError(err)
		}

		bookmark.Code = code
		return tx.WithContext(ctx).Model(bookmark).Update("code", code).Error
	})
}

// Update updates an existing bookmark for a specific user, only updating non-nil fields.
// Returns the number of rows affected and any error encountered.
func (r *repository) Update(ctx context.Context, id, userID string, updates *model.Bookmark) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(updates)

	if result.Error != nil {
		return 0, dbutils.ClassifyError(result.Error)
	}

	return result.RowsAffected, nil
}

// Delete deletes a bookmark for a specific user.
// Returns the number of rows affected and any error encountered.
func (r *repository) Delete(ctx context.Context, id, userID string) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		Delete(&model.Bookmark{})

	if result.Error != nil {
		return 0, dbutils.ClassifyError(result.Error)
	}

	return result.RowsAffected, nil
}
