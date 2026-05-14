package service

import (
	"context"
	"errors"
	"testing"

	"github.com/huypham67/bookmark-management/internal/dto/request"
	repoMocks "github.com/huypham67/bookmark-management/internal/repository/mocks"
	utilsMocks "github.com/huypham67/bookmark-management/internal/utils/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLinkService_ShortenURL(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		args       request.ShortenURLRequest
		setupMocks func(
			*repoMocks.Link,
			*utilsMocks.CodeGenerator,
		)
		verifyResponse func(*testing.T, string, error)
	}{
		{
			name: "should shorten URL successfully when code does not exist",
			args: request.ShortenURLRequest{
				Url: "https://google.com",
				Exp: 3600,
			},
			setupMocks: func(mockRepo *repoMocks.Link, mockCodeGen *utilsMocks.CodeGenerator) {
				mockCodeGen.
					On("Generate", shortCodeLength).
					Return("abc1234", nil).
					Once()

				mockRepo.
					On("CheckExists", mock.Anything, "abc1234").
					Return(false, nil).
					Once()

				mockRepo.
					On(
						"SaveLink",
						mock.Anything,
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
			args: request.ShortenURLRequest{
				Url: "https://google.com",
				Exp: 3600,
			},
			setupMocks: func(mockRepo *repoMocks.Link, mockCodeGen *utilsMocks.CodeGenerator) {
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
			args: request.ShortenURLRequest{
				Url: "https://google.com",
				Exp: 3600,
			},
			setupMocks: func(mockRepo *repoMocks.Link, mockCodeGen *utilsMocks.CodeGenerator) {
				mockCodeGen.
					On("Generate", shortCodeLength).
					Return("abc1234", nil).
					Once()

				mockRepo.
					On("CheckExists", mock.Anything, "abc1234").
					Return(false, errors.New("redis error")).
					Once()
			},
			verifyResponse: func(t *testing.T, code string, err error) {
				assert.Error(t, err)
				assert.Empty(t, code)
			},
		}, {
			name: "should retry code generation when code already exists",
			args: request.ShortenURLRequest{
				Url: "https://google.com",
				Exp: 3600,
			},
			setupMocks: func(mockRepo *repoMocks.Link, mockCodeGen *utilsMocks.CodeGenerator) {
				// First attempt
				mockCodeGen.
					On("Generate", shortCodeLength).
					Return("abc1234", nil).
					Once()

				mockRepo.
					On("CheckExists", mock.Anything, "abc1234").
					Return(true, nil).
					Once()

				// Second attempt
				mockCodeGen.
					On("Generate", shortCodeLength).
					Return("def5678", nil).
					Once()

				mockRepo.
					On("CheckExists", mock.Anything, "def5678").
					Return(false, nil).
					Once()

				mockRepo.
					On(
						"SaveLink",
						mock.Anything,
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
			args: request.ShortenURLRequest{
				Url: "https://google.com",
				Exp: 3600,
			},
			setupMocks: func(mockRepo *repoMocks.Link, mockCodeGen *utilsMocks.CodeGenerator) {
				mockCodeGen.
					On("Generate", shortCodeLength).
					Return("abc1234", nil).
					Once()

				mockRepo.
					On("CheckExists", mock.Anything, "abc1234").
					Return(false, nil).
					Once()

				mockRepo.
					On(
						"SaveLink",
						mock.Anything,
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

			mockRepo := new(repoMocks.Link)
			mockCodeGen := new(utilsMocks.CodeGenerator)

			tc.setupMocks(mockRepo, mockCodeGen)

			service := NewLinkService(mockRepo, mockCodeGen)

			ctx := context.Background()
			code, err := service.ShortenURL(ctx, tc.args)

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
		setupMocks     func(*repoMocks.Link)
		verifyResponse func(*testing.T, string, error)
	}{
		{
			name: "should get original URL successfully",
			args: args{
				code: "abc1234",
			},
			setupMocks: func(mockRepo *repoMocks.Link) {
				mockRepo.
					On("GetLink", mock.Anything, "abc1234").
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
			setupMocks: func(mockRepo *repoMocks.Link) {
				mockRepo.
					On("GetLink", mock.Anything, "missing").
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
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := new(repoMocks.Link)
			mockCodeGen := new(utilsMocks.CodeGenerator)
			tc.setupMocks(mockRepo)

			service := NewLinkService(mockRepo, mockCodeGen)

			ctx := context.Background()
			url, err := service.GetOriginalURL(ctx, tc.args.code)

			tc.verifyResponse(t, url, err)

			mockRepo.AssertExpectations(t)
			mockCodeGen.AssertExpectations(t)
		})
	}
}
