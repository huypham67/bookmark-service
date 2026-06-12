package cache

import (
	"context"
	"testing"

	"github.com/huypham67/bookmark-service/internal/repository/cache/mocks"
	bookmarkSvc "github.com/huypham67/bookmark-service/internal/service/bookmark"
	bookmarkMocks "github.com/huypham67/bookmark-service/internal/service/bookmark/mocks"
	"github.com/stretchr/testify/assert"
)

func TestService_Delete(t *testing.T) {
	t.Parallel()

	type args struct {
		userID     string
		bookmarkID string
	}

	testCases := []struct {
		name           string
		args           args
		setupMocks     func(context.Context, string, string, *bookmarkMocks.Service, *mocks.Repository)
		verifyResponse func(*testing.T, error)
	}{
		{
			name: "should delete bookmark and invalidate cache",
			args: args{
				userID:     "user-123",
				bookmarkID: "bm-1",
			},
			setupMocks: func(ctx context.Context, userID, bookmarkID string, bookmarkServiceMock *bookmarkMocks.Service, cacheRepoMock *mocks.Repository) {
				bookmarkServiceMock.On(
					"Delete",
					ctx,
					userID,
					bookmarkID,
				).Return(nil).Once()

				cacheRepoMock.On(
					"DeleteCacheByHashKey",
					ctx,
					"bookmarks:user-123",
				).Return(nil).Once()
			},
			verifyResponse: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "should return error when bookmark service fails",
			args: args{
				userID:     "user-456",
				bookmarkID: "bm-2",
			},
			setupMocks: func(ctx context.Context, userID, bookmarkID string, bookmarkServiceMock *bookmarkMocks.Service, cacheRepoMock *mocks.Repository) {
				// Cache is invalidated first; the DB write then fails.
				cacheRepoMock.On(
					"DeleteCacheByHashKey",
					ctx,
					"bookmarks:user-456",
				).Return(nil).Once()

				bookmarkServiceMock.On(
					"Delete",
					ctx,
					userID,
					bookmarkID,
				).Return(assert.AnError).Once()
			},
			verifyResponse: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "should not write and return error when cache invalidation fails",
			args: args{
				userID:     "user-789",
				bookmarkID: "bm-3",
			},
			setupMocks: func(ctx context.Context, userID, bookmarkID string, bookmarkServiceMock *bookmarkMocks.Service, cacheRepoMock *mocks.Repository) {
				// Cache invalidation fails, so the bookmark service must NOT be called.
				cacheRepoMock.On(
					"DeleteCacheByHashKey",
					ctx,
					"bookmarks:user-789",
				).Return(assert.AnError).Once()
			},
			verifyResponse: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, bookmarkSvc.ErrInternalServerError)
			},
		},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			bookmarkServiceMock := bookmarkMocks.NewService(t)
			cacheRepoMock := mocks.NewRepository(t)

			tc.setupMocks(ctx, tc.args.userID, tc.args.bookmarkID, bookmarkServiceMock, cacheRepoMock)

			service := NewBookmarkService(bookmarkServiceMock, cacheRepoMock)

			err := service.Delete(ctx, tc.args.userID, tc.args.bookmarkID)

			tc.verifyResponse(t, err)

			bookmarkServiceMock.AssertExpectations(t)
			cacheRepoMock.AssertExpectations(t)
		})
	}
}
