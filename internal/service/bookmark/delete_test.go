package bookmark

import (
	"context"
	"testing"

	"github.com/huypham67/bookmark-common/pkg/dbutils"
	bookmarkMocks "github.com/huypham67/bookmark-service/internal/repository/bookmark/mocks"
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
		setupMocks     func(context.Context, *bookmarkMocks.Repository)
		verifyResponse func(*testing.T, error)
	}{
		{
			name: "should delete bookmark successfully",
			args: args{
				userID:     "user-id-123",
				bookmarkID: "bookmark-id-456",
			},
			setupMocks: func(
				ctx context.Context,
				bookmarkRepo *bookmarkMocks.Repository,
			) {
				bookmarkRepo.
					On("Delete", ctx, "bookmark-id-456", "user-id-123").
					Return(int64(1), nil).
					Once()
			},
			verifyResponse: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "should return error when bookmark not found",
			args: args{
				userID:     "user-id-123",
				bookmarkID: "nonexistent-bookmark",
			},
			setupMocks: func(
				ctx context.Context,
				bookmarkRepo *bookmarkMocks.Repository,
			) {
				bookmarkRepo.
					On("Delete", ctx, "nonexistent-bookmark", "user-id-123").
					Return(int64(0), nil).
					Once()
			},
			verifyResponse: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrBookmarkNotFound)
			},
		},
		{
			name: "should return error when user not found",
			args: args{
				userID:     "nonexistent-user",
				bookmarkID: "bookmark-id-456",
			},
			setupMocks: func(
				ctx context.Context,
				bookmarkRepo *bookmarkMocks.Repository,
			) {
				bookmarkRepo.
					On("Delete", ctx, "bookmark-id-456", "nonexistent-user").
					Return(int64(0), dbutils.ErrForeignKeyViolationType).
					Once()
			},
			verifyResponse: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrBadRequest)
			},
		},
		{
			name: "should return error when database operation fails",
			args: args{
				userID:     "user-id-123",
				bookmarkID: "bookmark-id-456",
			},
			setupMocks: func(
				ctx context.Context,
				bookmarkRepo *bookmarkMocks.Repository,
			) {
				bookmarkRepo.
					On("Delete", ctx, "bookmark-id-456", "user-id-123").
					Return(int64(0), assert.AnError).
					Once()
			},
			verifyResponse: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInternalServerError)
			},
		},
		{
			name: "should return error when context is cancelled",
			args: args{
				userID:     "user-id-123",
				bookmarkID: "bookmark-id-456",
			},
			setupMocks: func(
				ctx context.Context,
				bookmarkRepo *bookmarkMocks.Repository,
			) {
				bookmarkRepo.
					On("Delete", ctx, "bookmark-id-456", "user-id-123").
					Return(int64(0), context.Canceled).
					Once()
			},
			verifyResponse: func(t *testing.T, err error) {
				assert.Error(t, err)
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

			err := bookmarkService.Delete(ctx, tc.args.userID, tc.args.bookmarkID)

			tc.verifyResponse(t, err)

			bookmarkRepo.AssertExpectations(t)
		})
	}
}
