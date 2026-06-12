package cache

import (
	"context"
	"testing"

	bookmarkDTO "github.com/huypham67/bookmark-service/internal/dto/bookmark"
	"github.com/huypham67/bookmark-service/internal/model"
	"github.com/huypham67/bookmark-service/internal/repository/cache/mocks"
	bookmarkSvc "github.com/huypham67/bookmark-service/internal/service/bookmark"
	bookmarkMocks "github.com/huypham67/bookmark-service/internal/service/bookmark/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_Create(t *testing.T) {
	t.Parallel()

	type args struct {
		userID  string
		request bookmarkDTO.CreateBookmarkRequest
	}

	testCases := []struct {
		name           string
		args           args
		setupMocks     func(context.Context, *bookmarkDTO.CreateBookmarkRequest, *bookmarkMocks.Service, *mocks.Repository)
		verifyResponse func(*testing.T, *model.Bookmark, error)
	}{
		{
			name: "should create bookmark and invalidate cache",
			args: args{
				userID: "user-123",
				request: bookmarkDTO.CreateBookmarkRequest{
					URL:         "https://example.com",
					Description: "Example",
				},
			},
			setupMocks: func(ctx context.Context, req *bookmarkDTO.CreateBookmarkRequest, bookmarkServiceMock *bookmarkMocks.Service, cacheRepoMock *mocks.Repository) {
				bookmark := &model.Bookmark{
					BaseModel: model.BaseModel{ID: "bm-1"},
					URL:       "https://example.com",
					Code:      "abc123",
					UserID:    "user-123",
				}

				bookmarkServiceMock.On(
					"Create",
					ctx,
					"user-123",
					*req,
				).Return(bookmark, nil).Once()

				cacheRepoMock.On(
					"DeleteCacheByHashKey",
					ctx,
					"bookmarks:user-123",
				).Return(nil).Once()
			},
			verifyResponse: func(t *testing.T, bm *model.Bookmark, err error) {
				assert.NoError(t, err)
				require.NotNil(t, bm)
				assert.Equal(t, "bm-1", bm.ID)
			},
		},
		{
			name: "should return error when bookmark service fails",
			args: args{
				userID: "user-456",
				request: bookmarkDTO.CreateBookmarkRequest{
					URL:         "https://example2.com",
					Description: "Example 2",
				},
			},
			setupMocks: func(ctx context.Context, req *bookmarkDTO.CreateBookmarkRequest, bookmarkServiceMock *bookmarkMocks.Service, cacheRepoMock *mocks.Repository) {
				// Cache is invalidated first; the DB write then fails.
				cacheRepoMock.On(
					"DeleteCacheByHashKey",
					ctx,
					"bookmarks:user-456",
				).Return(nil).Once()

				bookmarkServiceMock.On(
					"Create",
					ctx,
					"user-456",
					*req,
				).Return(nil, assert.AnError).Once()
			},
			verifyResponse: func(t *testing.T, bm *model.Bookmark, err error) {
				assert.Error(t, err)
				assert.Nil(t, bm)
			},
		},
		{
			name: "should not write and return error when cache invalidation fails",
			args: args{
				userID: "user-789",
				request: bookmarkDTO.CreateBookmarkRequest{
					URL:         "https://example3.com",
					Description: "Example 3",
				},
			},
			setupMocks: func(ctx context.Context, req *bookmarkDTO.CreateBookmarkRequest, bookmarkServiceMock *bookmarkMocks.Service, cacheRepoMock *mocks.Repository) {
				// Cache invalidation fails, so the bookmark service must NOT be called.
				cacheRepoMock.On(
					"DeleteCacheByHashKey",
					ctx,
					"bookmarks:user-789",
				).Return(assert.AnError).Once()
			},
			verifyResponse: func(t *testing.T, bm *model.Bookmark, err error) {
				assert.Nil(t, bm)
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

			tc.setupMocks(ctx, &tc.args.request, bookmarkServiceMock, cacheRepoMock)

			service := NewBookmarkService(bookmarkServiceMock, cacheRepoMock)

			bm, err := service.Create(ctx, tc.args.userID, tc.args.request)

			tc.verifyResponse(t, bm, err)

			bookmarkServiceMock.AssertExpectations(t)
			cacheRepoMock.AssertExpectations(t)
		})
	}
}
