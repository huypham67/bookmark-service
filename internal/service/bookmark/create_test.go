package bookmark

import (
	"context"
	"strings"
	"testing"

	"github.com/huypham67/bookmark-common/pkg/base62"
	"github.com/huypham67/bookmark-common/pkg/dbutils"
	"github.com/huypham67/bookmark-common/pkg/shortcode"
	bookmarkDTO "github.com/huypham67/bookmark-service/internal/dto/bookmark"
	"github.com/huypham67/bookmark-service/internal/model"
	bookmarkMocks "github.com/huypham67/bookmark-service/internal/repository/bookmark/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// matchBookmark verifies the bookmark fields and that Code is a valid
// SQL-routed code for the given code_int: <sqlPrefix> + base62(codeInt).
func matchBookmark(expected *model.Bookmark, codeInt int64) interface{} {
	return mock.MatchedBy(func(actual *model.Bookmark) bool {
		if actual.Description != expected.Description ||
			actual.URL != expected.URL ||
			actual.UserID != expected.UserID ||
			actual.CodeInt != int(codeInt) {
			return false
		}
		if len(actual.Code) < 2 {
			return false
		}
		prefixOK := strings.IndexByte(shortcode.SQLPrefixes, actual.Code[0]) >= 0
		payloadOK := actual.Code[1:] == base62.FormatUint(uint64(codeInt))
		return prefixOK && payloadOK
	})
}

func TestService_Create(t *testing.T) {
	t.Parallel()

	const codeInt int64 = 42

	type args struct {
		userID  string
		request bookmarkDTO.CreateBookmarkRequest
	}

	baseReq := bookmarkDTO.CreateBookmarkRequest{
		Description: "Test Bookmark",
		URL:         "https://example.com",
	}

	testCases := []struct {
		name           string
		args           args
		setupMocks     func(context.Context, *bookmarkMocks.Repository)
		verifyResponse func(*testing.T, *model.Bookmark, error)
	}{
		{
			name: "should create bookmark successfully",
			args: args{userID: "user-id-123", request: baseReq},
			setupMocks: func(ctx context.Context, bookmarkRepo *bookmarkMocks.Repository) {
				bookmarkRepo.On("NextCodeInt", ctx).Return(codeInt, nil).Once()

				expected := &model.Bookmark{
					Description: "Test Bookmark",
					URL:         "https://example.com",
					UserID:      "user-id-123",
				}
				bookmarkRepo.
					On("Create", ctx, matchBookmark(expected, codeInt)).
					Return(nil).
					Once()
			},
			verifyResponse: func(t *testing.T, bm *model.Bookmark, err error) {
				assert.NoError(t, err)
				require.NotNil(t, bm)
				assert.Equal(t, "Test Bookmark", bm.Description)
				assert.Equal(t, "https://example.com", bm.URL)
				assert.Equal(t, "user-id-123", bm.UserID)
				assert.Equal(t, int(codeInt), bm.CodeInt)
				assert.Equal(t, shortcode.StoreSQL, shortcode.Classify(bm.Code))
				assert.Equal(t, base62.FormatUint(uint64(codeInt)), bm.Code[1:])
			},
		},
		{
			name: "should return error when reserving code_int fails",
			args: args{userID: "user-id-123", request: baseReq},
			setupMocks: func(ctx context.Context, bookmarkRepo *bookmarkMocks.Repository) {
				bookmarkRepo.On("NextCodeInt", ctx).Return(int64(0), assert.AnError).Once()
			},
			verifyResponse: func(t *testing.T, bm *model.Bookmark, err error) {
				assert.Nil(t, bm)
				assert.ErrorIs(t, err, ErrInternalServerError)
			},
		},
		{
			name: "should return error when bookmark code already exists",
			args: args{userID: "user-id-123", request: baseReq},
			setupMocks: func(ctx context.Context, bookmarkRepo *bookmarkMocks.Repository) {
				bookmarkRepo.On("NextCodeInt", ctx).Return(codeInt, nil).Once()

				expected := &model.Bookmark{
					Description: "Test Bookmark",
					URL:         "https://example.com",
					UserID:      "user-id-123",
				}
				bookmarkRepo.
					On("Create", ctx, matchBookmark(expected, codeInt)).
					Return(dbutils.ErrDuplicationType).
					Once()
			},
			verifyResponse: func(t *testing.T, bm *model.Bookmark, err error) {
				assert.Nil(t, bm)
				assert.ErrorIs(t, err, ErrBookmarkAlreadyExists)
			},
		},
		{
			name: "should return error when user not found",
			args: args{userID: "nonexistent-user", request: baseReq},
			setupMocks: func(ctx context.Context, bookmarkRepo *bookmarkMocks.Repository) {
				bookmarkRepo.On("NextCodeInt", ctx).Return(codeInt, nil).Once()

				expected := &model.Bookmark{
					Description: "Test Bookmark",
					URL:         "https://example.com",
					UserID:      "nonexistent-user",
				}
				bookmarkRepo.
					On("Create", ctx, matchBookmark(expected, codeInt)).
					Return(dbutils.ErrForeignKeyViolationType).
					Once()
			},
			verifyResponse: func(t *testing.T, bm *model.Bookmark, err error) {
				assert.Nil(t, bm)
				assert.ErrorIs(t, err, ErrBadRequest)
			},
		},
		{
			name: "should return error when database operation fails",
			args: args{userID: "user-id-123", request: baseReq},
			setupMocks: func(ctx context.Context, bookmarkRepo *bookmarkMocks.Repository) {
				bookmarkRepo.On("NextCodeInt", ctx).Return(codeInt, nil).Once()

				expected := &model.Bookmark{
					Description: "Test Bookmark",
					URL:         "https://example.com",
					UserID:      "user-id-123",
				}
				bookmarkRepo.
					On("Create", ctx, matchBookmark(expected, codeInt)).
					Return(assert.AnError).
					Once()
			},
			verifyResponse: func(t *testing.T, bm *model.Bookmark, err error) {
				assert.Nil(t, bm)
				assert.ErrorIs(t, err, ErrInternalServerError)
			},
		},
		{
			name: "should return error when context is cancelled",
			args: args{userID: "user-id-123", request: baseReq},
			setupMocks: func(ctx context.Context, bookmarkRepo *bookmarkMocks.Repository) {
				bookmarkRepo.On("NextCodeInt", ctx).Return(int64(0), context.Canceled).Once()
			},
			verifyResponse: func(t *testing.T, bm *model.Bookmark, err error) {
				assert.Nil(t, bm)
				assert.ErrorIs(t, err, ErrInternalServerError)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var ctx context.Context
			bookmarkRepo := bookmarkMocks.NewRepository(t)

			if tc.name == "should return error when context is cancelled" {
				cancelledCtx, cancel := context.WithCancel(context.Background())
				cancel()
				ctx = cancelledCtx
			} else {
				ctx = context.Background()
			}

			tc.setupMocks(ctx, bookmarkRepo)

			bookmarkService := NewService(bookmarkRepo)

			bm, err := bookmarkService.Create(ctx, tc.args.userID, tc.args.request)

			tc.verifyResponse(t, bm, err)

			bookmarkRepo.AssertExpectations(t)
		})
	}
}
