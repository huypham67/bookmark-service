package bookmark

import (
	"context"

	"github.com/huypham67/bookmark-common/pkg/dbutils"
	"github.com/huypham67/bookmark-service/internal/model"
)

// Create saves a new bookmark to the database.
func (r *repository) Create(ctx context.Context, bookmark *model.Bookmark) error {
	if err := r.db.WithContext(ctx).Create(bookmark).Error; err != nil {
		return dbutils.ClassifyError(err)
	}
	return nil
}

// NextCodeInt reserves the next code_int value before a row is inserted.
//
// The bookmark's code is derived from code_int, so we need the integer up front
// (code is required at INSERT time, but code_int is normally
// assigned by the DB during the INSERT). The two branches below are not
// interchangeable styling — they reflect a real difference between the engines:
//
//   - Postgres (prod): code_int is SERIAL, backed by a real sequence, so
//     nextval() hands out the next value atomically and is safe under
//     concurrent creates (two callers can never get the same number).
//   - SQLite (tests only): code_int is not the INTEGER PRIMARY KEY, so SQLite
//     has no sequence for it; MAX(code_int)+1 is the only portable option. It
//     is NOT concurrency-safe, which is acceptable because tests run serially.
func (r *repository) NextCodeInt(ctx context.Context) (int64, error) {
	query := "SELECT COALESCE(MAX(code_int), 0) + 1 FROM bookmarks"
	if r.db.Dialector.Name() == "postgres" {
		query = "SELECT nextval(pg_get_serial_sequence('bookmarks', 'code_int'))"
	}

	var n int64
	if err := r.db.WithContext(ctx).Raw(query).Scan(&n).Error; err != nil {
		return 0, dbutils.ClassifyError(err)
	}
	return n, nil
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
