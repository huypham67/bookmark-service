package service

import (
	"context"
	"errors"
	"testing"

	"github.com/huypham67/bookmark-service/internal/dto/request"
	repoMocks "github.com/huypham67/bookmark-service/internal/repository/mocks"
	utilsMocks "github.com/huypham67/bookmark-service/pkg/utils/mocks"
	"github.com/stretchr/testify/assert"
)

func TestLinkService_ShortenURL(t *testing.T) {
	t.Parallel()

	type args struct {
		request request.ShortenURLRequest
	}

	testCases := []struct {
		name           string
		args           args
		setupMocks     func(context.Context, *repoMocks.Link, *utilsMocks.CodeGenerator)
		verifyResponse func(*testing.T, string, error)
	}{
		{
			name: "should shorten URL successfully when code does not exist",
			args: args{
				request: request.ShortenURLRequest{
					Url: "https://google.com",
					Exp: 3600,
				},
			},
			setupMocks: func(ctx context.Context, mockRepo *repoMocks.Link, mockCodeGen *utilsMocks.CodeGenerator) {
				mockCodeGen.
					On("Generate", shortCodeLength).
					Return("abc1234", nil).
					Once()

				mockRepo.
					On("CheckExists", ctx, "abc1234").
					Return(false, nil).
					Once()

				mockRepo.
					On(
						"SaveLink",
						ctx,
						"abc1234",
						"https://google.com",
						int64(3600),
					).
					Return(nil).
					Once()
			},
			verifyResponse: func(t *testing.T, code string, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "abc1234", code)
			},
		},
		{
			name: "should return error when code generation fails",
			args: args{
				request: request.ShortenURLRequest{
					Url: "https://google.com",
					Exp: 3600,
				},
			},
			setupMocks: func(ctx context.Context, mockRepo *repoMocks.Link, mockCodeGen *utilsMocks.CodeGenerator) {
				mockCodeGen.
					On("Generate", shortCodeLength).
					Return("", errors.New("code generation failed")).
					Once()
			},
			verifyResponse: func(t *testing.T, code string, err error) {
				assert.Error(t, err)
				assert.Empty(t, code)
			},
		}, {
			name: "should return error when checking code existence fails",
			args: args{
				request: request.ShortenURLRequest{
					Url: "https://google.com",
					Exp: 3600,
				},
			},
			setupMocks: func(ctx context.Context, mockRepo *repoMocks.Link, mockCodeGen *utilsMocks.CodeGenerator) {
				mockCodeGen.
					On("Generate", shortCodeLength).
					Return("abc1234", nil).
					Once()

				mockRepo.
					On("CheckExists", ctx, "abc1234").
					Return(false, errors.New("redis error")).
					Once()
			},
			verifyResponse: func(t *testing.T, code string, err error) {
				assert.Error(t, err)
				assert.Empty(t, code)
			},
		}, {
			name: "should retry code generation when code already exists",
			args: args{
				request: request.ShortenURLRequest{
					Url: "https://google.com",
					Exp: 3600,
				},
			},
			setupMocks: func(ctx context.Context, mockRepo *repoMocks.Link, mockCodeGen *utilsMocks.CodeGenerator) {
				// First attempt
				mockCodeGen.
					On("Generate", shortCodeLength).
					Return("abc1234", nil).
					Once()

				mockRepo.
					On("CheckExists", ctx, "abc1234").
					Return(true, nil).
					Once()

				// Second attempt
				mockCodeGen.
					On("Generate", shortCodeLength).
					Return("def5678", nil).
					Once()

				mockRepo.
					On("CheckExists", ctx, "def5678").
					Return(false, nil).
					Once()

				mockRepo.
					On(
						"SaveLink",
						ctx,
						"def5678",
						"https://google.com",
						int64(3600),
					).
					Return(nil).
					Once()
			},
			verifyResponse: func(t *testing.T, code string, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "def5678", code)
			},
		}, {
			name: "should return error when saving link fails",
			args: args{
				request: request.ShortenURLRequest{
					Url: "https://google.com",
					Exp: 3600,
				},
			},
			setupMocks: func(ctx context.Context, mockRepo *repoMocks.Link, mockCodeGen *utilsMocks.CodeGenerator) {
				mockCodeGen.
					On("Generate", shortCodeLength).
					Return("abc1234", nil).
					Once()

				mockRepo.
					On("CheckExists", ctx, "abc1234").
					Return(false, nil).
					Once()

				mockRepo.
					On(
						"SaveLink",
						ctx,
						"abc1234",
						"https://google.com",
						int64(3600),
					).
					Return(errors.New("save error")).
					Once()
			},
			verifyResponse: func(t *testing.T, code string, err error) {
				assert.Error(t, err)
				assert.Empty(t, code)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			mockRepo := new(repoMocks.Link)
			mockCodeGen := new(utilsMocks.CodeGenerator)

			tc.setupMocks(ctx, mockRepo, mockCodeGen)

			service := NewLinkService(mockRepo, mockCodeGen)

			code, err := service.ShortenURL(ctx, tc.args.request)

			tc.verifyResponse(t, code, err)

			mockRepo.AssertExpectations(t)
			mockCodeGen.AssertExpectations(t)
		})
	}
}

func TestLinkService_GetOriginalURL(t *testing.T) {
	t.Parallel()

	type args struct {
		code string
	}

	testCases := []struct {
		name           string
		args           args
		setupMocks     func(context.Context, *repoMocks.Link)
		verifyResponse func(*testing.T, string, error)
	}{
		{
			name: "should get original URL successfully",
			args: args{
				code: "abc1234",
			},
			setupMocks: func(ctx context.Context, mockRepo *repoMocks.Link) {
				mockRepo.
					On("GetLink", ctx, "abc1234").
					Return("https://google.com", nil).
					Once()
			},
			verifyResponse: func(t *testing.T, url string, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "https://google.com", url)
			},
		},
		{
			name: "should return error when code does not exist",
			args: args{
				code: "missing",
			},
			setupMocks: func(ctx context.Context, mockRepo *repoMocks.Link) {
				mockRepo.
					On("GetLink", ctx, "missing").
					Return("", errors.New("code not found")).
					Once()
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

			ctx := context.Background()
			mockRepo := new(repoMocks.Link)
			mockCodeGen := new(utilsMocks.CodeGenerator)
			tc.setupMocks(ctx, mockRepo)

			service := NewLinkService(mockRepo, mockCodeGen)

			url, err := service.GetOriginalURL(ctx, tc.args.code)

			tc.verifyResponse(t, url, err)

			mockRepo.AssertExpectations(t)
			mockCodeGen.AssertExpectations(t)
		})
	}
}
