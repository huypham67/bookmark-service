package bookmark

import (
	"context"
	"testing"

	bookmarkDTO "github.com/huypham67/bookmark-service/internal/dto/bookmark"
	"github.com/huypham67/bookmark-service/internal/model"
	bookmarkMocks "github.com/huypham67/bookmark-service/internal/repository/bookmark/mocks"
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
		setupMocks     func(context.Context, *bookmarkMocks.Repository)
		verifyResponse func(*testing.T, []*model.Bookmark, *PaginationResult, error)
	}{
		{
			name: "should list bookmarks successfully",
			args: args{
				userID: "user-id-123",
				request: &bookmarkDTO.ListBookmarksRequest{
					Page:  1,
					Limit: 10,
					Sort:  "created_at",
				},
			},
			setupMocks: func(
				ctx context.Context,
				bookmarkRepo *bookmarkMocks.Repository,
			) {
				bookmarks := []*model.Bookmark{
					{
						BaseModel: model.BaseModel{
							ID: "bm-1",
						},
						Description: "Bookmark 1",
						URL:         "https://example1.com",
						Code:        "abc123",
						UserID:      "user-id-123",
					},
					{
						BaseModel: model.BaseModel{
							ID: "bm-2",
						},
						Description: "Bookmark 2",
						URL:         "https://example2.com",
						Code:        "def456",
						UserID:      "user-id-123",
					},
				}

				bookmarkRepo.
					On("GetPaginatedByUserID", ctx, "user-id-123", int64(0), int64(10), "created_at").
					Return(bookmarks, nil).
					Once()

				bookmarkRepo.
					On("CountByUserID", ctx, "user-id-123").
					Return(int64(2), nil).
					Once()
			},
			verifyResponse: func(t *testing.T, bms []*model.Bookmark, pr *PaginationResult, err error) {
				assert.NoError(t, err)
				require.NotNil(t, bms)
				require.NotNil(t, pr)

				assert.Len(t, bms, 2)
				assert.Equal(t, "bm-1", bms[0].ID)
				assert.Equal(t, "bm-2", bms[1].ID)

				assert.Equal(t, int64(1), pr.Page)
				assert.Equal(t, int64(10), pr.Limit)
				assert.Equal(t, int64(2), pr.Total)
			},
		},
		{
			name: "should list bookmarks with empty result",
			args: args{
				userID: "user-id-456",
				request: &bookmarkDTO.ListBookmarksRequest{
					Page:  1,
					Limit: 10,
					Sort:  "updated_at",
				},
			},
			setupMocks: func(
				ctx context.Context,
				bookmarkRepo *bookmarkMocks.Repository,
			) {
				emptyBookmarks := make([]*model.Bookmark, 0)

				bookmarkRepo.
					On("GetPaginatedByUserID", ctx, "user-id-456", int64(0), int64(10), "updated_at").
					Return(emptyBookmarks, nil).
					Once()

				bookmarkRepo.
					On("CountByUserID", ctx, "user-id-456").
					Return(int64(0), nil).
					Once()
			},
			verifyResponse: func(t *testing.T, bms []*model.Bookmark, pr *PaginationResult, err error) {
				assert.NoError(t, err)
				require.NotNil(t, bms)
				require.NotNil(t, pr)

				assert.Empty(t, bms)
				assert.Equal(t, int64(1), pr.Page)
				assert.Equal(t, int64(10), pr.Limit)
				assert.Equal(t, int64(0), pr.Total)
			},
		},
		{
			name: "should return error when GetPaginatedByUserID fails",
			args: args{
				userID: "user-id-123",
				request: &bookmarkDTO.ListBookmarksRequest{
					Page:  2,
					Limit: 20,
					Sort:  "code",
				},
			},
			setupMocks: func(
				ctx context.Context,
				bookmarkRepo *bookmarkMocks.Repository,
			) {
				bookmarkRepo.
					On("GetPaginatedByUserID", ctx, "user-id-123", int64(20), int64(20), "code").
					Return(nil, assert.AnError).
					Once()
			},
			verifyResponse: func(t *testing.T, bms []*model.Bookmark, pr *PaginationResult, err error) {
				assert.Error(t, err)
				assert.Nil(t, bms)
				assert.Nil(t, pr)
				assert.ErrorIs(t, err, ErrInternalServerError)
			},
		},
		{
			name: "should return error when CountByUserID fails",
			args: args{
				userID: "user-id-123",
				request: &bookmarkDTO.ListBookmarksRequest{
					Page:  1,
					Limit: 10,
					Sort:  "url",
				},
			},
			setupMocks: func(
				ctx context.Context,
				bookmarkRepo *bookmarkMocks.Repository,
			) {
				bookmarks := []*model.Bookmark{
					{
						BaseModel: model.BaseModel{
							ID: "bm-1",
						},
						Description: "Bookmark 1",
						URL:         "https://example1.com",
						Code:        "abc123",
						UserID:      "user-id-123",
					},
				}

				bookmarkRepo.
					On("GetPaginatedByUserID", ctx, "user-id-123", int64(0), int64(10), "url").
					Return(bookmarks, nil).
					Once()

				bookmarkRepo.
					On("CountByUserID", ctx, "user-id-123").
					Return(int64(0), assert.AnError).
					Once()
			},
			verifyResponse: func(t *testing.T, bms []*model.Bookmark, pr *PaginationResult, err error) {
				assert.Error(t, err)
				assert.Nil(t, bms)
				assert.Nil(t, pr)
				assert.ErrorIs(t, err, ErrInternalServerError)
			},
		},
		{
			name: "should return error when context is cancelled",
			args: args{
				userID: "user-id-123",
				request: &bookmarkDTO.ListBookmarksRequest{
					Page:  1,
					Limit: 10,
					Sort:  "created_at",
				},
			},
			setupMocks: func(
				ctx context.Context,
				bookmarkRepo *bookmarkMocks.Repository,
			) {
				bookmarkRepo.
					On("GetPaginatedByUserID", ctx, "user-id-123", int64(0), int64(10), "created_at").
					Return(nil, context.Canceled).
					Once()
			},
			verifyResponse: func(t *testing.T, bms []*model.Bookmark, pr *PaginationResult, err error) {
				assert.Error(t, err)
				assert.Nil(t, bms)
				assert.Nil(t, pr)
				assert.ErrorIs(t, err, ErrInternalServerError)
			},
		},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var ctx context.Context
			bookmarkRepo := bookmarkMocks.NewRepository(t)

			// For context cancellation test, create a cancelled context
			if tc.name == "should return error when context is cancelled" {
				cancelledCtx, cancel := context.WithCancel(context.Background())
				cancel()
				ctx = cancelledCtx
			} else {
				ctx = context.Background()
			}

			tc.setupMocks(ctx, bookmarkRepo)

			bookmarkService := NewService(bookmarkRepo)

			bms, pr, err := bookmarkService.List(ctx, tc.args.userID, tc.args.request)

			tc.verifyResponse(t, bms, pr, err)

			bookmarkRepo.AssertExpectations(t)
		})
	}
}
