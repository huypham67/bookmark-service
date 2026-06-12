package link

import (
	"context"
	"testing"

	"github.com/huypham67/bookmark-common/pkg/dbutils"
	utilsMocks "github.com/huypham67/bookmark-common/pkg/utils/mocks"
	linkRepoMocks "github.com/huypham67/bookmark-service/internal/repository/link/mocks"
	resolverMocks "github.com/huypham67/bookmark-service/internal/service/link/resolver/mocks"
	"github.com/stretchr/testify/assert"
)

func TestService_GetOriginalURL(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		code           string
		setupMocks     func(context.Context, *linkRepoMocks.Repository, *resolverMocks.Bookmark)
		verifyResponse func(*testing.T, string, error)
	}{
		{
			name: "should resolve redis code from Redis store",
			code: "abc1234", // prefix 'a' -> Redis
			setupMocks: func(ctx context.Context, repo *linkRepoMocks.Repository, resolver *resolverMocks.Bookmark) {
				repo.On("GetLink", ctx, "abc1234").Return("https://google.com", nil).Once()
			},
			verifyResponse: func(t *testing.T, url string, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "https://google.com", url)
			},
		},
		{
			name: "should resolve sql code from bookmark store",
			code: "qg", // prefix 'q' -> SQL
			setupMocks: func(ctx context.Context, repo *linkRepoMocks.Repository, resolver *resolverMocks.Bookmark) {
				resolver.On("GetURLByCode", ctx, "qg").Return("https://bookmark.com", nil).Once()
			},
			verifyResponse: func(t *testing.T, url string, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "https://bookmark.com", url)
			},
		},
		{
			name: "should return error when redis code not found",
			code: "amissing", // prefix 'a' -> Redis
			setupMocks: func(ctx context.Context, repo *linkRepoMocks.Repository, resolver *resolverMocks.Bookmark) {
				repo.On("GetLink", ctx, "amissing").Return("", dbutils.ErrRecordNotFoundType).Once()
			},
			verifyResponse: func(t *testing.T, url string, err error) {
				assert.ErrorIs(t, err, dbutils.ErrRecordNotFoundType)
				assert.Empty(t, url)
			},
		},
		{
			name: "should return error when sql code not found",
			code: "missing", // prefix 'm' -> SQL
			setupMocks: func(ctx context.Context, repo *linkRepoMocks.Repository, resolver *resolverMocks.Bookmark) {
				resolver.On("GetURLByCode", ctx, "missing").Return("", dbutils.ErrRecordNotFoundType).Once()
			},
			verifyResponse: func(t *testing.T, url string, err error) {
				assert.ErrorIs(t, err, dbutils.ErrRecordNotFoundType)
				assert.Empty(t, url)
			},
		},
		{
			name: "should return not found when prefix is unrecognised",
			code: "5xyz", // prefix '5' -> neither bucket
			setupMocks: func(ctx context.Context, repo *linkRepoMocks.Repository, resolver *resolverMocks.Bookmark) {
				// no store should be queried
			},
			verifyResponse: func(t *testing.T, url string, err error) {
				assert.ErrorIs(t, err, dbutils.ErrRecordNotFoundType)
				assert.Empty(t, url)
			},
		},
		{
			name: "should return error when context is cancelled",
			code: "abc1234",
			setupMocks: func(ctx context.Context, repo *linkRepoMocks.Repository, resolver *resolverMocks.Bookmark) {
				repo.On("GetLink", ctx, "abc1234").Return("", context.Canceled).Once()
			},
			verifyResponse: func(t *testing.T, url string, err error) {
				assert.Error(t, err)
				assert.Empty(t, url)
			},
		},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var ctx context.Context
			mockRepo := linkRepoMocks.NewRepository(t)
			mockCodeGen := new(utilsMocks.CodeGenerator)
			mockResolver := resolverMocks.NewBookmark(t)

			if tc.name == "should return error when context is cancelled" {
				cancelledCtx, cancel := context.WithCancel(context.Background())
				cancel()
				ctx = cancelledCtx
			} else {
				ctx = context.Background()
			}

			tc.setupMocks(ctx, mockRepo, mockResolver)

			service := NewService(mockRepo, mockCodeGen, mockResolver)

			url, err := service.GetOriginalURL(ctx, tc.code)

			tc.verifyResponse(t, url, err)

			mockRepo.AssertExpectations(t)
			mockResolver.AssertExpectations(t)
		})
	}
}
