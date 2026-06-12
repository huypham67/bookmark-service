package bookmark

import (
	"context"
	"testing"

	"github.com/huypham67/bookmark-service/internal/model"
	"github.com/huypham67/bookmark-service/internal/test/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepository_GetPaginatedByUserID(t *testing.T) {
	t.Parallel()

	type args struct {
		userID string
		offset int64
		limit  int64
		sort   string
	}

	testCases := []struct {
		name   string
		args   args
		verify func(*testing.T, []*model.Bookmark, error)
	}{
		{
			name: "should return first page of bookmarks",
			args: args{
				userID: "user-uuid-1",
				offset: 0,
				limit:  3,
				sort:   "created_at",
			},
			verify: func(t *testing.T, bookmarks []*model.Bookmark, err error) {
				require.NoError(t, err)
				require.NotNil(t, bookmarks)
				assert.Len(t, bookmarks, 3)
				for _, bm := range bookmarks {
					assert.Equal(t, "user-uuid-1", bm.UserID)
					assert.NotEmpty(t, bm.ID)
				}
				// Verify DESC order: first should be most recent
				assert.True(t, bookmarks[0].CreatedAt.After(bookmarks[2].CreatedAt))
			},
		},
		{
			name: "should return middle page of bookmarks",
			args: args{
				userID: "user-uuid-1",
				offset: 3,
				limit:  3,
				sort:   "created_at",
			},
			verify: func(t *testing.T, bookmarks []*model.Bookmark, err error) {
				require.NoError(t, err)
				require.NotNil(t, bookmarks)
				assert.Len(t, bookmarks, 3)
				for _, bm := range bookmarks {
					assert.Equal(t, "user-uuid-1", bm.UserID)
				}
				// Verify DESC order
				assert.True(t, bookmarks[0].CreatedAt.After(bookmarks[2].CreatedAt))
			},
		},
		{
			name: "should return last page with partial results",
			args: args{
				userID: "user-uuid-1",
				offset: 6,
				limit:  3,
				sort:   "created_at",
			},
			verify: func(t *testing.T, bookmarks []*model.Bookmark, err error) {
				require.NoError(t, err)
				require.NotNil(t, bookmarks)
				assert.Len(t, bookmarks, 2)
				for _, bm := range bookmarks {
					assert.Equal(t, "user-uuid-1", bm.UserID)
				}
				// Verify DESC order
				assert.True(t, bookmarks[0].CreatedAt.After(bookmarks[1].CreatedAt))
			},
		},
		{
			name: "should return bookmarks for different user",
			args: args{
				userID: "user-uuid-2",
				offset: 0,
				limit:  10,
				sort:   "created_at",
			},
			verify: func(t *testing.T, bookmarks []*model.Bookmark, err error) {
				require.NoError(t, err)
				require.NotNil(t, bookmarks)
				assert.Len(t, bookmarks, 5)
				for _, bm := range bookmarks {
					assert.Equal(t, "user-uuid-2", bm.UserID)
				}
			},
		},
		{
			name: "should return bookmarks sorted by updated_at",
			args: args{
				userID: "user-uuid-1",
				offset: 0,
				limit:  8,
				sort:   "updated_at",
			},
			verify: func(t *testing.T, bookmarks []*model.Bookmark, err error) {
				require.NoError(t, err)
				require.NotNil(t, bookmarks)
				assert.Len(t, bookmarks, 8)
				for _, bm := range bookmarks {
					assert.Equal(t, "user-uuid-1", bm.UserID)
				}
			},
		},
		{
			name: "should return empty when offset is beyond total count",
			args: args{
				userID: "user-uuid-1",
				offset: 100,
				limit:  10,
				sort:   "created_at",
			},
			verify: func(t *testing.T, bookmarks []*model.Bookmark, err error) {
				require.NoError(t, err)
				require.NotNil(t, bookmarks)
				assert.Len(t, bookmarks, 0)
			},
		},
		{
			name: "should return empty for user with no bookmarks",
			args: args{
				userID: "user-uuid-3",
				offset: 0,
				limit:  10,
				sort:   "created_at",
			},
			verify: func(t *testing.T, bookmarks []*model.Bookmark, err error) {
				require.NoError(t, err)
				require.NotNil(t, bookmarks)
				assert.Len(t, bookmarks, 0)
			},
		},
		{
			name: "should return empty when limit is zero",
			args: args{
				userID: "user-uuid-1",
				offset: 0,
				limit:  0,
				sort:   "created_at",
			},
			verify: func(t *testing.T, bookmarks []*model.Bookmark, err error) {
				require.NoError(t, err)
				require.NotNil(t, bookmarks)
				assert.Len(t, bookmarks, 0)
			},
		},
		{
			name: "should return empty for nonexistent user",
			args: args{
				userID: "nonexistent-user-id",
				offset: 0,
				limit:  10,
				sort:   "created_at",
			},
			verify: func(t *testing.T, bookmarks []*model.Bookmark, err error) {
				require.NoError(t, err)
				require.NotNil(t, bookmarks)
				assert.Len(t, bookmarks, 0)
			},
		},
		{
			name: "should return all bookmarks with large limit",
			args: args{
				userID: "user-uuid-1",
				offset: 0,
				limit:  1000,
				sort:   "created_at",
			},
			verify: func(t *testing.T, bookmarks []*model.Bookmark, err error) {
				require.NoError(t, err)
				require.NotNil(t, bookmarks)
				assert.Len(t, bookmarks, 8)
				for _, bm := range bookmarks {
					assert.Equal(t, "user-uuid-1", bm.UserID)
				}
				// Verify DESC order: first should be most recent, last should be oldest
				assert.True(t, bookmarks[0].CreatedAt.After(bookmarks[7].CreatedAt))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			repo, _ := newTestRepository(t)

			bookmarks, err := repo.GetPaginatedByUserID(ctx, tc.args.userID, tc.args.offset, tc.args.limit, tc.args.sort)

			tc.verify(t, bookmarks, err)
		})
	}
}

func TestRepository_CountByUserID(t *testing.T) {
	t.Parallel()

	type args struct {
		userID string
	}

	testCases := []struct {
		name   string
		args   args
		verify func(*testing.T, int64, error)
	}{
		{
			name: "should return count for user with multiple bookmarks",
			args: args{
				userID: fixtures.TestUserID1,
			},
			verify: func(t *testing.T, count int64, err error) {
				require.NoError(t, err)
				assert.Equal(t, int64(8), count)
			},
		},
		{
			name: "should return count for user with fewer bookmarks",
			args: args{
				userID: fixtures.TestUserID2,
			},
			verify: func(t *testing.T, count int64, err error) {
				require.NoError(t, err)
				assert.Equal(t, int64(5), count)
			},
		},
		{
			name: "should return zero for user with no bookmarks",
			args: args{
				userID: "user-uuid-3",
			},
			verify: func(t *testing.T, count int64, err error) {
				require.NoError(t, err)
				assert.Equal(t, int64(0), count)
			},
		},
		{
			name: "should return zero for nonexistent user",
			args: args{
				userID: "nonexistent-user-id",
			},
			verify: func(t *testing.T, count int64, err error) {
				require.NoError(t, err)
				assert.Equal(t, int64(0), count)
			},
		},
		{
			name: "should return error when context is cancelled",
			args: args{
				userID: fixtures.TestUserID1,
			},
			verify: func(t *testing.T, count int64, err error) {
				require.Error(t, err)
				assert.Equal(t, int64(0), count)
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

			repo, _ := newTestRepository(t)

			count, err := repo.CountByUserID(ctx, tc.args.userID)

			tc.verify(t, count, err)
		})
	}
}
