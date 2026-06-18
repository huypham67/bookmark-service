package bookmark

import (
	"context"
	"testing"

	"github.com/huypham67/bookmark-common/pkg/dbutils"
	"github.com/huypham67/bookmark-common/pkg/shortcode"
	"github.com/huypham67/bookmark-service/internal/model"
	"github.com/huypham67/bookmark-service/internal/test/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestRepository_Create(t *testing.T) {
	t.Parallel()

	type args struct {
		bookmark *model.Bookmark
	}

	testCases := []struct {
		name   string
		args   args
		verify func(*testing.T, *gorm.DB, error, args)
	}{
		{
			name: "should create bookmark successfully",
			args: args{
				bookmark: &model.Bookmark{
					BaseModel:   model.BaseModel{ID: "new-bookmark-1"},
					Description: "New Test Bookmark",
					URL:         "https://example.com/new",
					UserID:      fixtures.TestUserID1,
				},
			},
			verify: func(t *testing.T, db *gorm.DB, err error, a args) {
				require.NoError(t, err)

				var actual model.Bookmark
				result := db.First(&actual, "id = ?", a.bookmark.ID)
				require.NoError(t, result.Error)

				assert.NotEmpty(t, actual.Code)
				assert.Equal(t, a.bookmark.Code, actual.Code)
				assert.Equal(t, shortcode.StoreSQL, shortcode.Classify(actual.Code))
				assert.Equal(t, a.bookmark.URL, actual.URL)
				assert.Equal(t, a.bookmark.UserID, actual.UserID)
				assert.Equal(t, a.bookmark.Description, actual.Description)
			},
		},
		{
			name: "should create multiple bookmarks for same user",
			args: args{
				bookmark: &model.Bookmark{
					BaseModel:   model.BaseModel{ID: "new-bookmark-2"},
					Description: "Another New Bookmark",
					URL:         "https://example.com/another",
					UserID:      fixtures.TestUserID1,
				},
			},
			verify: func(t *testing.T, db *gorm.DB, err error, a args) {
				require.NoError(t, err)

				var actual model.Bookmark
				result := db.First(&actual, "id = ?", a.bookmark.ID)
				require.NoError(t, result.Error)

				assert.NotEmpty(t, actual.Code)
				assert.Equal(t, a.bookmark.Code, actual.Code)
				assert.Equal(t, a.bookmark.UserID, actual.UserID)
			},
		},
		{
			name: "should return error when ID is duplicate",
			args: args{
				bookmark: &model.Bookmark{
					BaseModel:   model.BaseModel{ID: "bookmark-1-1"}, // already exists in seed
					Description: "Duplicate ID Bookmark",
					URL:         "https://example.com/duplicate",
					UserID:      fixtures.TestUserID2,
				},
			},
			verify: func(t *testing.T, db *gorm.DB, err error, a args) {
				require.Error(t, err)
				assert.ErrorIs(t, err, dbutils.ErrDuplicationType)
			},
		},
		{
			name: "should create bookmark with minimal fields",
			args: args{
				bookmark: &model.Bookmark{
					BaseModel:   model.BaseModel{ID: "minimal-bookmark"},
					URL:         "https://example.com/minimal",
					UserID:      fixtures.TestUserID2,
					Description: "",
				},
			},
			verify: func(t *testing.T, db *gorm.DB, err error, a args) {
				require.NoError(t, err)

				var actual model.Bookmark
				result := db.First(&actual, "id = ?", a.bookmark.ID)
				require.NoError(t, result.Error)

				assert.NotEmpty(t, actual.Code)
				assert.Equal(t, a.bookmark.URL, actual.URL)
			},
		},
		{
			name: "should return error when context is cancelled",
			args: args{
				bookmark: &model.Bookmark{
					BaseModel:   model.BaseModel{ID: "cancel-bookmark"},
					Description: "Cancelled Context Bookmark",
					URL:         "https://example.com/cancel",
					UserID:      fixtures.TestUserID1,
				},
			},
			verify: func(t *testing.T, db *gorm.DB, err error, a args) {
				require.Error(t, err)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			// For context cancellation test, create cancelled context
			if tc.name == "should return error when context is cancelled" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			repo, db := newTestRepository(t)

			err := repo.Create(ctx, tc.args.bookmark)

			tc.verify(t, db, err, tc.args)
		})
	}
}

func TestRepository_Update(t *testing.T) {
	t.Parallel()

	type args struct {
		id      string
		userID  string
		updates *model.Bookmark
	}

	testCases := []struct {
		name   string
		args   args
		verify func(*testing.T, *gorm.DB, int64, error)
	}{
		{
			name: "should update bookmark successfully",
			args: args{
				id:     "bookmark-1-1",
				userID: fixtures.TestUserID1,
				updates: &model.Bookmark{
					Description: "Updated Description",
					URL:         "https://example.com/updated",
				},
			},
			verify: func(t *testing.T, db *gorm.DB, rowsAffected int64, err error) {
				require.NoError(t, err)
				assert.Equal(t, int64(1), rowsAffected)

				var actual model.Bookmark
				result := db.First(&actual, "id = ?", "bookmark-1-1")
				require.NoError(t, result.Error)

				assert.Equal(t, "Updated Description", actual.Description)
				assert.Equal(t, "https://example.com/updated", actual.URL)
				assert.Equal(t, "code1001", actual.Code)
			},
		},
		{
			name: "should update partial fields",
			args: args{
				id:     "bookmark-1-2",
				userID: fixtures.TestUserID1,
				updates: &model.Bookmark{
					Description: "Only Description Changed",
				},
			},
			verify: func(t *testing.T, db *gorm.DB, rowsAffected int64, err error) {
				require.NoError(t, err)
				assert.Equal(t, int64(1), rowsAffected)

				var actual model.Bookmark
				result := db.First(&actual, "id = ?", "bookmark-1-2")
				require.NoError(t, result.Error)

				assert.Equal(t, "Only Description Changed", actual.Description)
				assert.Equal(t, "https://example.com/2", actual.URL) // unchanged
			},
		},
		{
			name: "should return zero rows affected when bookmark does not exist",
			args: args{
				id:     "nonexistent-bookmark",
				userID: fixtures.TestUserID1,
				updates: &model.Bookmark{
					Description: "Updated",
					URL:         "https://example.com/updated",
				},
			},
			verify: func(t *testing.T, db *gorm.DB, rowsAffected int64, err error) {
				require.NoError(t, err)
				assert.Equal(t, int64(0), rowsAffected)
			},
		},
		{
			name: "should return zero rows affected when userID does not match (security check)",
			args: args{
				id:     "bookmark-1-1",
				userID: fixtures.TestUserID2, // different user
				updates: &model.Bookmark{
					Description: "Updated",
					URL:         "https://example.com/updated",
				},
			},
			verify: func(t *testing.T, db *gorm.DB, rowsAffected int64, err error) {
				require.NoError(t, err)
				assert.Equal(t, int64(0), rowsAffected)

				var actual model.Bookmark
				result := db.First(&actual, "id = ?", "bookmark-1-1")
				require.NoError(t, result.Error)
				assert.Equal(t, "Test Bookmark 1-1", actual.Description) // unchanged
			},
		},
		{
			name: "should return error when context is cancelled",
			args: args{
				id:     "bookmark-1-1",
				userID: fixtures.TestUserID1,
				updates: &model.Bookmark{
					Description: "Updated",
				},
			},
			verify: func(t *testing.T, db *gorm.DB, rowsAffected int64, err error) {
				require.Error(t, err)
				assert.Equal(t, int64(0), rowsAffected)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			// For context cancellation test, create cancelled context
			if tc.name == "should return error when context is cancelled" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			repo, db := newTestRepository(t)

			rowsAffected, err := repo.Update(ctx, tc.args.id, tc.args.userID, tc.args.updates)

			tc.verify(t, db, rowsAffected, err)
		})
	}
}

func TestRepository_Delete(t *testing.T) {
	t.Parallel()

	type args struct {
		id     string
		userID string
	}

	testCases := []struct {
		name   string
		args   args
		verify func(*testing.T, *gorm.DB, int64, error)
	}{
		{
			name: "should delete bookmark successfully",
			args: args{
				id:     "bookmark-1-1",
				userID: fixtures.TestUserID1,
			},
			verify: func(t *testing.T, db *gorm.DB, rowsAffected int64, err error) {
				require.NoError(t, err)
				assert.Equal(t, int64(1), rowsAffected)

				var actual model.Bookmark
				result := db.First(&actual, "id = ?", "bookmark-1-1")
				assert.ErrorIs(t, result.Error, gorm.ErrRecordNotFound)
			},
		},
		{
			name: "should return zero rows affected when bookmark does not exist",
			args: args{
				id:     "nonexistent-bookmark",
				userID: fixtures.TestUserID1,
			},
			verify: func(t *testing.T, db *gorm.DB, rowsAffected int64, err error) {
				require.NoError(t, err)
				assert.Equal(t, int64(0), rowsAffected)
			},
		},
		{
			name: "should return zero rows affected when userID does not match (security check)",
			args: args{
				id:     "bookmark-1-2",
				userID: fixtures.TestUserID2, // different user
			},
			verify: func(t *testing.T, db *gorm.DB, rowsAffected int64, err error) {
				require.NoError(t, err)
				assert.Equal(t, int64(0), rowsAffected)

				var actual model.Bookmark
				result := db.First(&actual, "id = ?", "bookmark-1-2")
				require.NoError(t, result.Error)
				assert.Equal(t, "Test Bookmark 1-2", actual.Description) // still exists
			},
		},
		{
			name: "should return error when context is cancelled",
			args: args{
				id:     "bookmark-1-3",
				userID: fixtures.TestUserID1,
			},
			verify: func(t *testing.T, db *gorm.DB, rowsAffected int64, err error) {
				require.Error(t, err)
				assert.Equal(t, int64(0), rowsAffected)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			// For context cancellation test, create cancelled context
			if tc.name == "should return error when context is cancelled" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			repo, db := newTestRepository(t)

			rowsAffected, err := repo.Delete(ctx, tc.args.id, tc.args.userID)

			tc.verify(t, db, rowsAffected, err)
		})
	}
}
