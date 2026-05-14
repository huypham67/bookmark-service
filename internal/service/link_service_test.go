package service

import (
	"errors"
	"testing"

	"github.com/huypham67/bookmark-management/internal/dto/request"
	repoMocks "github.com/huypham67/bookmark-management/internal/repository/mocks"
	utilsMocks "github.com/huypham67/bookmark-management/internal/utils/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLinkService_ShortenURL(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		requestURL    string
		requestExp    int64
		generatedCode string
		setupMocks    func(*repoMocks.Link, *utilsMocks.CodeGenerator)
		expectError   bool
	}{
		{
			name:          "should shorten URL successfully when code does not exist",
			requestURL:    "https://example.com/very/long/path",
			requestExp:    3600,
			generatedCode: "abc1234",
			setupMocks: func(repo *repoMocks.Link, cg *utilsMocks.CodeGenerator) {
				cg.On("Generate", 7).Return("abc1234", nil).Once()
				repo.On("CheckExists", "abc1234").Return(false, nil).Once()
				repo.On("SaveLink", "abc1234", "https://example.com/very/long/path", int64(3600)).Return(nil).Once()
			},
			expectError: false,
		},
		{
			name:          "should return error when GenerateCode fails",
			requestURL:    "https://example.com",
			requestExp:    3600,
			generatedCode: "",
			setupMocks: func(repo *repoMocks.Link, cg *utilsMocks.CodeGenerator) {
				cg.On("Generate", 7).Return("", errors.New("random generation failed")).Once()
			},
			expectError: true,
		},
		{
			name:          "should return error when CheckExists fails",
			requestURL:    "https://example.com",
			requestExp:    3600,
			generatedCode: "def5678",
			setupMocks: func(repo *repoMocks.Link, cg *utilsMocks.CodeGenerator) {
				cg.On("Generate", 7).Return("def5678", nil).Once()
				repo.On("CheckExists", "def5678").Return(false, errors.New("redis connection error")).Once()
			},
			expectError: true,
		},
		{
			name:          "should return error when SaveLink fails",
			requestURL:    "https://example.com",
			requestExp:    3600,
			generatedCode: "ghi9012",
			setupMocks: func(repo *repoMocks.Link, cg *utilsMocks.CodeGenerator) {
				cg.On("Generate", 7).Return("ghi9012", nil).Once()
				repo.On("CheckExists", "ghi9012").Return(false, nil).Once()
				repo.On("SaveLink", "ghi9012", "https://example.com", int64(3600)).Return(errors.New("save failed")).Once()
			},
			expectError: true,
		},
		{
			name:          "should shorten URL with short expiration time",
			requestURL:    "https://google.com",
			requestExp:    60,
			generatedCode: "jkl3456",
			setupMocks: func(repo *repoMocks.Link, cg *utilsMocks.CodeGenerator) {
				cg.On("Generate", 7).Return("jkl3456", nil).Once()
				repo.On("CheckExists", "jkl3456").Return(false, nil).Once()
				repo.On("SaveLink", "jkl3456", "https://google.com", int64(60)).Return(nil).Once()
			},
			expectError: false,
		},
		{
			name:          "should shorten URL with zero expiration (no expiration)",
			requestURL:    "https://github.com",
			requestExp:    0,
			generatedCode: "mno7890",
			setupMocks: func(repo *repoMocks.Link, cg *utilsMocks.CodeGenerator) {
				cg.On("Generate", 7).Return("mno7890", nil).Once()
				repo.On("CheckExists", "mno7890").Return(false, nil).Once()
				repo.On("SaveLink", "mno7890", "https://github.com", int64(0)).Return(nil).Once()
			},
			expectError: false,
		},
		{
			name:          "should shorten URL with empty URL",
			requestURL:    "",
			requestExp:    3600,
			generatedCode: "pqr1357",
			setupMocks: func(repo *repoMocks.Link, cg *utilsMocks.CodeGenerator) {
				cg.On("Generate", 7).Return("pqr1357", nil).Once()
				repo.On("CheckExists", "pqr1357").Return(false, nil).Once()
				repo.On("SaveLink", "pqr1357", "", int64(3600)).Return(nil).Once()
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := repoMocks.NewLink(t)
			mockCodeGen := utilsMocks.NewCodeGenerator(t)
			tc.setupMocks(mockRepo, mockCodeGen)

			service := NewLinkService(mockRepo, mockCodeGen)

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
			assert.Equal(t, tc.generatedCode, code)
		})
	}
}

func TestLinkService_ShortenURL_RetryOnCodeConflict(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		requestURL     string
		requestExp     int64
		generatedCodes []string
		setupMocks     func(*repoMocks.Link, *utilsMocks.CodeGenerator)
		expectError    bool
	}{
		{
			name:           "should retry when generated code already exists once",
			requestURL:     "https://example.com",
			requestExp:     3600,
			generatedCodes: []string{"code001", "code002"},
			setupMocks: func(repo *repoMocks.Link, cg *utilsMocks.CodeGenerator) {
				// First attempt: generate code001
				cg.On("Generate", 7).Return("code001", nil).Once()
				// code001 already exists
				repo.On("CheckExists", "code001").Return(true, nil).Once()

				// Second attempt: generate code002
				cg.On("Generate", 7).Return("code002", nil).Once()
				// code002 does not exist
				repo.On("CheckExists", "code002").Return(false, nil).Once()
				// Save code002
				repo.On("SaveLink", "code002", "https://example.com", int64(3600)).Return(nil).Once()
			},
			expectError: false,
		},
		{
			name:           "should retry multiple times when code conflicts occur",
			requestURL:     "https://example.com",
			requestExp:     3600,
			generatedCodes: []string{"code001", "code002", "code003"},
			setupMocks: func(repo *repoMocks.Link, cg *utilsMocks.CodeGenerator) {
				// First attempt: generate code001
				cg.On("Generate", 7).Return("code001", nil).Once()
				// code001 already exists
				repo.On("CheckExists", "code001").Return(true, nil).Once()

				// Second attempt: generate code002
				cg.On("Generate", 7).Return("code002", nil).Once()
				// code002 already exists
				repo.On("CheckExists", "code002").Return(true, nil).Once()

				// Third attempt: generate code003
				cg.On("Generate", 7).Return("code003", nil).Once()
				// code003 does not exist
				repo.On("CheckExists", "code003").Return(false, nil).Once()
				// Save code003
				repo.On("SaveLink", "code003", "https://example.com", int64(3600)).Return(nil).Once()
			},
			expectError: false,
		},
		{
			name:           "should return error when retry fails due to CheckExists error",
			requestURL:     "https://example.com",
			requestExp:     3600,
			generatedCodes: []string{"code001", "code002"},
			setupMocks: func(repo *repoMocks.Link, cg *utilsMocks.CodeGenerator) {
				// First attempt: generate code001
				cg.On("Generate", 7).Return("code001", nil).Once()
				// code001 already exists
				repo.On("CheckExists", "code001").Return(true, nil).Once()

				// Second attempt: generate code002
				cg.On("Generate", 7).Return("code002", nil).Once()
				// CheckExists fails on retry
				repo.On("CheckExists", "code002").Return(false, errors.New("redis error on retry")).Once()
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := repoMocks.NewLink(t)
			mockCodeGen := utilsMocks.NewCodeGenerator(t)
			tc.setupMocks(mockRepo, mockCodeGen)

			service := NewLinkService(mockRepo, mockCodeGen)

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
				// Verify the returned code is one of the generated codes
				assert.Contains(t, tc.generatedCodes, code)
			}
		})
	}
}

func TestLinkService_NewLinkService(t *testing.T) {
	t.Parallel()

	mockRepo := repoMocks.NewLink(t)
	mockCodeGen := utilsMocks.NewCodeGenerator(t)

	service := NewLinkService(mockRepo, mockCodeGen)

	assert.NotNil(t, service)
	assert.Implements(t, (*LinkService)(nil), service)
}

func TestLinkService_ShortenURL_VerifyCodeFormat(t *testing.T) {
	t.Parallel()

	mockRepo := repoMocks.NewLink(t)
	mockCodeGen := utilsMocks.NewCodeGenerator(t)

	// Setup mocks with a code that contains only alphanumeric characters
	mockCodeGen.On("Generate", 7).Return("aBc1234", nil).Once()
	mockRepo.On("CheckExists", "aBc1234").Return(false, nil).Once()
	mockRepo.On("SaveLink", "aBc1234", "https://example.com", int64(3600)).Return(nil).Once()

	service := NewLinkService(mockRepo, mockCodeGen)

	req := request.ShortenURLRequest{
		Url: "https://example.com",
		Exp: 3600,
	}

	code, err := service.ShortenURL(req)

	require.NoError(t, err)
	require.NotEmpty(t, code)

	// Verify code format: should be 7 characters long
	assert.Equal(t, 7, len(code))
	assert.Equal(t, "aBc1234", code)

	// Verify code contains only alphanumeric characters
	for _, ch := range code {
		assert.True(t,
			(ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9'),
			"code should only contain alphanumeric characters")
	}
}

func TestLinkService_GetOriginalURL(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		code        string
		setupMocks  func(*repoMocks.Link)
		expectedURL string
		expectError bool
	}{
		{
			name: "should return original URL successfully",
			code: "abc1234",
			setupMocks: func(repo *repoMocks.Link) {
				repo.On("GetLink", "abc1234").Return("https://example.com/very/long/path", nil).Once()
			},
			expectedURL: "https://example.com/very/long/path",
			expectError: false,
		},
		{
			name: "should return error when link is not found",
			code: "notfound",
			setupMocks: func(repo *repoMocks.Link) {
				repo.On("GetLink", "notfound").Return("", errors.New("link not found")).Once()
			},
			expectedURL: "",
			expectError: true,
		},
		{
			name: "should return different URLs for different codes",
			code: "xyz7890",
			setupMocks: func(repo *repoMocks.Link) {
				repo.On("GetLink", "xyz7890").Return("https://github.com/huypham67/bookmark-management", nil).Once()
			},
			expectedURL: "https://github.com/huypham67/bookmark-management",
			expectError: false,
		},
		{
			name: "should handle URL with query parameters",
			code: "query01",
			setupMocks: func(repo *repoMocks.Link) {
				repo.On("GetLink", "query01").Return("https://example.com?key=value&foo=bar", nil).Once()
			},
			expectedURL: "https://example.com?key=value&foo=bar",
			expectError: false,
		},
		{
			name: "should handle URL with hash fragment",
			code: "hash001",
			setupMocks: func(repo *repoMocks.Link) {
				repo.On("GetLink", "hash001").Return("https://example.com#section", nil).Once()
			},
			expectedURL: "https://example.com#section",
			expectError: false,
		},
		{
			name: "should handle very long URL",
			code: "verylong",
			setupMocks: func(repo *repoMocks.Link) {
				repo.On("GetLink", "verylong").Return("https://example.com/very/long/path/with/many/segments?param1=value1&param2=value2&param3=value3", nil).Once()
			},
			expectedURL: "https://example.com/very/long/path/with/many/segments?param1=value1&param2=value2&param3=value3",
			expectError: false,
		},
		{
			name: "should handle redis connection error",
			code: "error01",
			setupMocks: func(repo *repoMocks.Link) {
				repo.On("GetLink", "error01").Return("", errors.New("redis connection error")).Once()
			},
			expectedURL: "",
			expectError: true,
		},
		{
			name: "should handle expired link",
			code: "expired",
			setupMocks: func(repo *repoMocks.Link) {
				repo.On("GetLink", "expired").Return("", errors.New("key expired")).Once()
			},
			expectedURL: "",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := repoMocks.NewLink(t)
			mockCodeGen := utilsMocks.NewCodeGenerator(t)
			tc.setupMocks(mockRepo)

			service := NewLinkService(mockRepo, mockCodeGen)

			url, err := service.GetOriginalURL(tc.code)

			if tc.expectError {
				assert.Error(t, err)
				assert.Empty(t, url)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedURL, url)
		})
	}
}
