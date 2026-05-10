package link

import (
	"errors"
	"testing"

	"github.com/huypham67/bookmark-management/internal/dto/request"
	"github.com/huypham67/bookmark-management/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestLinkService_ShortenURL(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		requestURL  string
		requestExp  int64
		setupMock   func(*mocks.LinkRepository)
		expectCode  string
		expectError bool
	}{
		{
			name:       "should shorten URL successfully when code does not exist",
			requestURL: "https://example.com/very/long/path",
			requestExp: 3600,
			setupMock: func(repo *mocks.LinkRepository) {
				repo.On("CheckExists", mock.MatchedBy(func(code string) bool {
					return len(code) == 7
				})).Return(false, nil).Once()
				repo.On("SaveLink", mock.MatchedBy(func(code string) bool {
					return len(code) == 7
				}), "https://example.com/very/long/path", int64(3600)).Return(nil).Once()
			},
			expectCode:  "", // We can't predict the random code
			expectError: false,
		},
		{
			name:       "should return error when CheckExists fails",
			requestURL: "https://example.com",
			requestExp: 3600,
			setupMock: func(repo *mocks.LinkRepository) {
				repo.On("CheckExists", mock.MatchedBy(func(code string) bool {
					return len(code) == 7
				})).Return(false, errors.New("redis connection error")).Once()
			},
			expectCode:  "",
			expectError: true,
		},
		{
			name:       "should return error when SaveLink fails",
			requestURL: "https://example.com",
			requestExp: 3600,
			setupMock: func(repo *mocks.LinkRepository) {
				repo.On("CheckExists", mock.MatchedBy(func(code string) bool {
					return len(code) == 7
				})).Return(false, nil).Once()
				repo.On("SaveLink", mock.MatchedBy(func(code string) bool {
					return len(code) == 7
				}), "https://example.com", int64(3600)).Return(errors.New("save failed")).Once()
			},
			expectCode:  "",
			expectError: true,
		},
		{
			name:       "should shorten URL with short expiration time",
			requestURL: "https://google.com",
			requestExp: 60,
			setupMock: func(repo *mocks.LinkRepository) {
				repo.On("CheckExists", mock.MatchedBy(func(code string) bool {
					return len(code) == 7
				})).Return(false, nil).Once()
				repo.On("SaveLink", mock.MatchedBy(func(code string) bool {
					return len(code) == 7
				}), "https://google.com", int64(60)).Return(nil).Once()
			},
			expectCode:  "",
			expectError: false,
		},
		{
			name:       "should shorten URL with zero expiration (no expiration)",
			requestURL: "https://github.com",
			requestExp: 0,
			setupMock: func(repo *mocks.LinkRepository) {
				repo.On("CheckExists", mock.MatchedBy(func(code string) bool {
					return len(code) == 7
				})).Return(false, nil).Once()
				repo.On("SaveLink", mock.MatchedBy(func(code string) bool {
					return len(code) == 7
				}), "https://github.com", int64(0)).Return(nil).Once()
			},
			expectCode:  "",
			expectError: false,
		},
		{
			name:       "should shorten URL with empty URL",
			requestURL: "",
			requestExp: 3600,
			setupMock: func(repo *mocks.LinkRepository) {
				repo.On("CheckExists", mock.MatchedBy(func(code string) bool {
					return len(code) == 7
				})).Return(false, nil).Once()
				repo.On("SaveLink", mock.MatchedBy(func(code string) bool {
					return len(code) == 7
				}), "", int64(3600)).Return(nil).Once()
			},
			expectCode:  "",
			expectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := mocks.NewLinkRepository(t)
			tc.setupMock(mockRepo)

			service := NewLinkService(mockRepo)

			req := request.ShortenURLRequest{
				Url: tc.requestURL,
				Exp: tc.requestExp,
			}

			code, err := service.ShortenURL(req)

			if tc.expectError {
				assert.Error(t, err)
				assert.Empty(t, code)
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, code)
			assert.Equal(t, 7, len(code))
		})
	}
}

func TestLinkService_ShortenURL_RetryOnCodeConflict(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		requestURL  string
		requestExp  int64
		setupMock   func(*mocks.LinkRepository)
		expectError bool
	}{
		{
			name:       "should retry when generated code already exists once",
			requestURL: "https://example.com",
			requestExp: 3600,
			setupMock: func(repo *mocks.LinkRepository) {
				// First attempt: code exists
				repo.On("CheckExists", mock.MatchedBy(func(code string) bool {
					return len(code) == 7
				})).Return(true, nil).Once()

				// Second attempt: code does not exist
				repo.On("CheckExists", mock.MatchedBy(func(code string) bool {
					return len(code) == 7
				})).Return(false, nil).Once()

				// Save link on second attempt
				repo.On("SaveLink", mock.MatchedBy(func(code string) bool {
					return len(code) == 7
				}), "https://example.com", int64(3600)).Return(nil).Once()
			},
			expectError: false,
		},
		{
			name:       "should retry multiple times when code conflicts occur",
			requestURL: "https://example.com",
			requestExp: 3600,
			setupMock: func(repo *mocks.LinkRepository) {
				// First attempt: code exists
				repo.On("CheckExists", mock.MatchedBy(func(code string) bool {
					return len(code) == 7
				})).Return(true, nil).Once()

				// Second attempt: code exists
				repo.On("CheckExists", mock.MatchedBy(func(code string) bool {
					return len(code) == 7
				})).Return(true, nil).Once()

				// Third attempt: code does not exist
				repo.On("CheckExists", mock.MatchedBy(func(code string) bool {
					return len(code) == 7
				})).Return(false, nil).Once()

				// Save link on third attempt
				repo.On("SaveLink", mock.MatchedBy(func(code string) bool {
					return len(code) == 7
				}), "https://example.com", int64(3600)).Return(nil).Once()
			},
			expectError: false,
		},
		{
			name:       "should return error when retry fails due to CheckExists error",
			requestURL: "https://example.com",
			requestExp: 3600,
			setupMock: func(repo *mocks.LinkRepository) {
				// First attempt: code exists
				repo.On("CheckExists", mock.MatchedBy(func(code string) bool {
					return len(code) == 7
				})).Return(true, nil).Once()

				// Second attempt: CheckExists fails
				repo.On("CheckExists", mock.MatchedBy(func(code string) bool {
					return len(code) == 7
				})).Return(false, errors.New("redis error on retry")).Once()
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := mocks.NewLinkRepository(t)
			tc.setupMock(mockRepo)

			service := NewLinkService(mockRepo)

			req := request.ShortenURLRequest{
				Url: tc.requestURL,
				Exp: tc.requestExp,
			}

			code, err := service.ShortenURL(req)

			if tc.expectError {
				assert.Error(t, err)
				assert.Empty(t, code)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, code)
				assert.Equal(t, 7, len(code))
			}
		})
	}
}

func TestLinkService_NewLinkService(t *testing.T) {
	t.Parallel()

	mockRepo := mocks.NewLinkRepository(t)

	service := NewLinkService(mockRepo)

	assert.NotNil(t, service)
	assert.Implements(t, (*Link)(nil), service)
}

func TestLinkService_ShortenURL_VerifyCodeFormat(t *testing.T) {
	t.Parallel()

	mockRepo := mocks.NewLinkRepository(t)
	mockRepo.On("CheckExists", mock.MatchedBy(func(code string) bool {
		return len(code) == 7
	})).Return(false, nil).Once()
	mockRepo.On("SaveLink", mock.MatchedBy(func(code string) bool {
		return len(code) == 7
	}), "https://example.com", int64(3600)).Return(nil).Once()

	service := NewLinkService(mockRepo)

	req := request.ShortenURLRequest{
		Url: "https://example.com",
		Exp: 3600,
	}

	code, err := service.ShortenURL(req)

	require.NoError(t, err)
	require.NotEmpty(t, code)

	// Verify code format: should be 7 characters long
	assert.Equal(t, 7, len(code))

	// Verify code contains only alphanumeric characters
	for _, ch := range code {
		assert.True(t,
			(ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9'),
			"code should only contain alphanumeric characters")
	}
}
