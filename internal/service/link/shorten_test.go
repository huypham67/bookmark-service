package link

import (
	"context"
	"errors"
	"testing"

	"github.com/huypham67/bookmark-common/pkg/shortcode"
	utilsMocks "github.com/huypham67/bookmark-common/pkg/utils/mocks"
	linkDTO "github.com/huypham67/bookmark-service/internal/dto/link"
	linkRepoMocks "github.com/huypham67/bookmark-service/internal/repository/link/mocks"
	resolverMocks "github.com/huypham67/bookmark-service/internal/service/link/resolver/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// redisCodeWithPayload matches a code that carries a Redis routing prefix
// followed by the given random payload.
func redisCodeWithPayload(payload string) interface{} {
	return mock.MatchedBy(func(code string) bool {
		return shortcode.Classify(code) == shortcode.StoreRedis && len(code) > 1 && code[1:] == payload
	})
}

func TestService_ShortenURL(t *testing.T) {
	t.Parallel()

	request := linkDTO.ShortenURLRequest{Url: "https://google.com", Exp: 3600}

	testCases := []struct {
		name           string
		setupMocks     func(context.Context, *linkRepoMocks.Repository, *utilsMocks.CodeGenerator)
		verifyResponse func(*testing.T, string, error)
	}{
		{
			name: "should shorten URL successfully when code does not exist",
			setupMocks: func(ctx context.Context, mockRepo *linkRepoMocks.Repository, mockCodeGen *utilsMocks.CodeGenerator) {
				mockCodeGen.On("Generate", shortCodeLength).Return("abc1234", nil).Once()
				mockRepo.On("CheckExists", ctx, redisCodeWithPayload("abc1234")).Return(false, nil).Once()
				mockRepo.On("SaveLink", ctx, redisCodeWithPayload("abc1234"), "https://google.com", int64(3600)).Return(nil).Once()
			},
			verifyResponse: func(t *testing.T, code string, err error) {
				assert.NoError(t, err)
				assert.Equal(t, shortcode.StoreRedis, shortcode.Classify(code))
				assert.Equal(t, "abc1234", code[1:])
			},
		},
		{
			name: "should return error when code generation fails",
			setupMocks: func(ctx context.Context, mockRepo *linkRepoMocks.Repository, mockCodeGen *utilsMocks.CodeGenerator) {
				mockCodeGen.On("Generate", shortCodeLength).Return("", errors.New("code generation failed")).Once()
			},
			verifyResponse: func(t *testing.T, code string, err error) {
				assert.Error(t, err)
				assert.Empty(t, code)
			},
		},
		{
			name: "should return error when checking code existence fails",
			setupMocks: func(ctx context.Context, mockRepo *linkRepoMocks.Repository, mockCodeGen *utilsMocks.CodeGenerator) {
				mockCodeGen.On("Generate", shortCodeLength).Return("abc1234", nil).Once()
				mockRepo.On("CheckExists", ctx, redisCodeWithPayload("abc1234")).Return(false, errors.New("redis error")).Once()
			},
			verifyResponse: func(t *testing.T, code string, err error) {
				assert.Error(t, err)
				assert.Empty(t, code)
			},
		},
		{
			name: "should retry code generation when code already exists",
			setupMocks: func(ctx context.Context, mockRepo *linkRepoMocks.Repository, mockCodeGen *utilsMocks.CodeGenerator) {
				mockCodeGen.On("Generate", shortCodeLength).Return("abc1234", nil).Once()
				mockRepo.On("CheckExists", ctx, redisCodeWithPayload("abc1234")).Return(true, nil).Once()

				mockCodeGen.On("Generate", shortCodeLength).Return("def5678", nil).Once()
				mockRepo.On("CheckExists", ctx, redisCodeWithPayload("def5678")).Return(false, nil).Once()
				mockRepo.On("SaveLink", ctx, redisCodeWithPayload("def5678"), "https://google.com", int64(3600)).Return(nil).Once()
			},
			verifyResponse: func(t *testing.T, code string, err error) {
				assert.NoError(t, err)
				assert.Equal(t, shortcode.StoreRedis, shortcode.Classify(code))
				assert.Equal(t, "def5678", code[1:])
			},
		},
		{
			name: "should return error when saving link fails",
			setupMocks: func(ctx context.Context, mockRepo *linkRepoMocks.Repository, mockCodeGen *utilsMocks.CodeGenerator) {
				mockCodeGen.On("Generate", shortCodeLength).Return("abc1234", nil).Once()
				mockRepo.On("CheckExists", ctx, redisCodeWithPayload("abc1234")).Return(false, nil).Once()
				mockRepo.On("SaveLink", ctx, redisCodeWithPayload("abc1234"), "https://google.com", int64(3600)).Return(errors.New("save error")).Once()
			},
			verifyResponse: func(t *testing.T, code string, err error) {
				assert.Error(t, err)
				assert.Empty(t, code)
			},
		},
		{
			name: "should return error when context is cancelled",
			setupMocks: func(ctx context.Context, mockRepo *linkRepoMocks.Repository, mockCodeGen *utilsMocks.CodeGenerator) {
				mockCodeGen.On("Generate", shortCodeLength).Return("abc1234", nil).Once()
				mockRepo.On("CheckExists", ctx, redisCodeWithPayload("abc1234")).Return(false, context.Canceled).Once()
			},
			verifyResponse: func(t *testing.T, code string, err error) {
				assert.Error(t, err)
				assert.Empty(t, code)
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

			tc.setupMocks(ctx, mockRepo, mockCodeGen)

			service := NewService(mockRepo, mockCodeGen, mockResolver)

			code, err := service.ShortenURL(ctx, request)

			tc.verifyResponse(t, code, err)

			mockRepo.AssertExpectations(t)
			mockCodeGen.AssertExpectations(t)
		})
	}
}
