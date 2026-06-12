package cache

import (
	"context"
	"encoding/json"
	"testing"

	bookmarkDTO "github.com/huypham67/bookmark-service/internal/dto/bookmark"
	"github.com/huypham67/bookmark-service/internal/model"
	"github.com/huypham67/bookmark-service/internal/repository/cache/mocks"
	"github.com/huypham67/bookmark-service/internal/service/bookmark"
	bookmarkMocks "github.com/huypham67/bookmark-service/internal/service/bookmark/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_List(t *testing.T) {
	t.Parallel()

	type args struct {
		userID  string
		request *bookmarkDTO.ListBookmarksRequest
	}

	testCases := []struct {
		name           string
		args           args
		setupMocks     func(context.Context, *bookmarkDTO.ListBookmarksRequest, *bookmarkMocks.Service, *mocks.Repository)
		verifyResponse func(*testing.T, []*model.Bookmark, *bookmark.PaginationResult, error)
	}{
		{
			name: "should return bookmarks from cache on hit",
			args: args{
				userID: "user-123",
				request: &bookmarkDTO.ListBookmarksRequest{
					Page:  1,
					Limit: 10,
					Sort:  "created_at",
				},
			},
			setupMocks: func(ctx context.Context, req *bookmarkDTO.ListBookmarksRequest, bookmarkServiceMock *bookmarkMocks.Service, cacheRepoMock *mocks.Repository) {
				bookmarks := []*model.Bookmark{
					{
						BaseModel: model.BaseModel{ID: "bm-1"},
						URL:       "https://example.com",
						Code:      "abc123",
						UserID:    "user-123",
					},
				}
				pagination := &bookmark.PaginationResult{
					Page:  1,
					Limit: 10,
					Total: 1,
				}
				cacheData := cacheData{
					Bookmarks:  bookmarks,
					Pagination: pagination,
				}
				cacheDataRaw, _ := json.Marshal(cacheData)

				cacheRepoMock.On(
					"GetCache",
					ctx,
					"bookmarks:user-123",
					"page:1:limit:10:sort:created_at",
				).Return(cacheDataRaw, nil).Once()
			},
			verifyResponse: func(t *testing.T, bms []*model.Bookmark, pr *bookmark.PaginationResult, err error) {
				assert.NoError(t, err)
				require.NotNil(t, bms)
				require.NotNil(t, pr)
				assert.Len(t, bms, 1)
				assert.Equal(t, "bm-1", bms[0].ID)
			},
		},
		{
			name: "should call bookmark service on cache miss and cache result",
			args: args{
				userID: "user-456",
				request: &bookmarkDTO.ListBookmarksRequest{
					Page:  1,
					Limit: 10,
					Sort:  "created_at",
				},
			},
			setupMocks: func(ctx context.Context, req *bookmarkDTO.ListBookmarksRequest, bookmarkServiceMock *bookmarkMocks.Service, cacheRepoMock *mocks.Repository) {
				bookmarks := []*model.Bookmark{
					{
						BaseModel: model.BaseModel{ID: "bm-2"},
						URL:       "https://example2.com",
						Code:      "def456",
						UserID:    "user-456",
					},
				}
				pagination := &bookmark.PaginationResult{
					Page:  1,
					Limit: 10,
					Total: 1,
				}

				cacheRepoMock.On(
					"GetCache",
					ctx,
					"bookmarks:user-456",
					"page:1:limit:10:sort:created_at",
				).Return(nil, assert.AnError).Once()

				bookmarkServiceMock.On(
					"List",
					ctx,
					"user-456",
					req,
				).Return(bookmarks, pagination, nil).Once()

				cacheData := cacheData{
					Bookmarks:  bookmarks,
					Pagination: pagination,
				}
				cacheDataBytes, _ := json.Marshal(cacheData)

				cacheRepoMock.On(
					"SetCache",
					ctx,
					"bookmarks:user-456",
					"page:1:limit:10:sort:created_at",
					cacheDataBytes,
					cacheTTL,
				).Return(nil).Once()
			},
			verifyResponse: func(t *testing.T, bms []*model.Bookmark, pr *bookmark.PaginationResult, err error) {
				assert.NoError(t, err)
				require.NotNil(t, bms)
				require.NotNil(t, pr)
				assert.Len(t, bms, 1)
				assert.Equal(t, "bm-2", bms[0].ID)
			},
		},
		{
			name: "should handle corrupted cache and call service",
			args: args{
				userID: "user-789",
				request: &bookmarkDTO.ListBookmarksRequest{
					Page:  1,
					Limit: 10,
					Sort:  "created_at",
				},
			},
			setupMocks: func(ctx context.Context, req *bookmarkDTO.ListBookmarksRequest, bookmarkServiceMock *bookmarkMocks.Service, cacheRepoMock *mocks.Repository) {
				bookmarks := []*model.Bookmark{
					{
						BaseModel: model.BaseModel{ID: "bm-3"},
						URL:       "https://example3.com",
						Code:      "ghi789",
						UserID:    "user-789",
					},
				}
				pagination := &bookmark.PaginationResult{
					Page:  1,
					Limit: 10,
					Total: 1,
				}

				cacheRepoMock.On(
					"GetCache",
					ctx,
					"bookmarks:user-789",
					"page:1:limit:10:sort:created_at",
				).Return([]byte("corrupted"), nil).Once()

				cacheRepoMock.On(
					"DeleteCacheByFieldKey",
					ctx,
					"bookmarks:user-789",
					"page:1:limit:10:sort:created_at",
				).Return(nil).Once()

				bookmarkServiceMock.On(
					"List",
					ctx,
					"user-789",
					req,
				).Return(bookmarks, pagination, nil).Once()

				cacheData := cacheData{
					Bookmarks:  bookmarks,
					Pagination: pagination,
				}
				cacheDataBytes, _ := json.Marshal(cacheData)

				cacheRepoMock.On(
					"SetCache",
					ctx,
					"bookmarks:user-789",
					"page:1:limit:10:sort:created_at",
					cacheDataBytes,
					cacheTTL,
				).Return(nil).Once()
			},
			verifyResponse: func(t *testing.T, bms []*model.Bookmark, pr *bookmark.PaginationResult, err error) {
				assert.NoError(t, err)
				require.NotNil(t, bms)
				assert.Len(t, bms, 1)
			},
		},
		{
			name: "should return error when bookmark service fails",
			args: args{
				userID: "user-999",
				request: &bookmarkDTO.ListBookmarksRequest{
					Page:  1,
					Limit: 10,
					Sort:  "created_at",
				},
			},
			setupMocks: func(ctx context.Context, req *bookmarkDTO.ListBookmarksRequest, bookmarkServiceMock *bookmarkMocks.Service, cacheRepoMock *mocks.Repository) {
				cacheRepoMock.On(
					"GetCache",
					ctx,
					"bookmarks:user-999",
					"page:1:limit:10:sort:created_at",
				).Return(nil, assert.AnError).Once()

				bookmarkServiceMock.On(
					"List",
					ctx,
					"user-999",
					req,
				).Return(nil, nil, assert.AnError).Once()
			},
			verifyResponse: func(t *testing.T, bms []*model.Bookmark, pr *bookmark.PaginationResult, err error) {
				assert.Error(t, err)
				assert.Nil(t, bms)
				assert.Nil(t, pr)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			bookmarkServiceMock := bookmarkMocks.NewService(t)
			cacheRepoMock := mocks.NewRepository(t)

			tc.setupMocks(ctx, tc.args.request, bookmarkServiceMock, cacheRepoMock)

			service := NewBookmarkService(bookmarkServiceMock, cacheRepoMock)

			bms, pr, err := service.List(ctx, tc.args.userID, tc.args.request)

			tc.verifyResponse(t, bms, pr, err)

			bookmarkServiceMock.AssertExpectations(t)
			cacheRepoMock.AssertExpectations(t)
		})
	}
}
