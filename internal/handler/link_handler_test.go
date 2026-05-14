package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/huypham67/bookmark-management/internal/dto/request"
	"github.com/huypham67/bookmark-management/internal/dto/response"
	"github.com/huypham67/bookmark-management/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLinkHandler_ShortenURL(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		requestBody     request.ShortenURLRequest
		setupMock       func(*mocks.LinkService)
		expectedCode    int
		expectedMessage string
		expectedCodeLen int
		expectError     bool
	}{
		{
			name: "should return 200 with shortened URL code",
			requestBody: request.ShortenURLRequest{
				Url: "https://example.com/very/long/path",
				Exp: 3600,
			},
			setupMock: func(svc *mocks.LinkService) {
				svc.On("ShortenURL", request.ShortenURLRequest{
					Url: "https://example.com/very/long/path",
					Exp: 3600,
				}).Return("abc123d", nil).Once()
			},
			expectedCode:    http.StatusOK,
			expectedMessage: "Shorten URL generated successfully",
			expectedCodeLen: 7,
			expectError:     false,
		},
		{
			name: "should return error when service fails",
			requestBody: request.ShortenURLRequest{
				Url: "https://example.com",
				Exp: 3600,
			},
			setupMock: func(svc *mocks.LinkService) {
				svc.On("ShortenURL", request.ShortenURLRequest{
					Url: "https://example.com",
					Exp: 3600,
				}).Return("", errors.New("redis error")).Once()
			},
			expectedCode:    http.StatusInternalServerError,
			expectedMessage: "Internal Server Error",
			expectedCodeLen: 0,
			expectError:     true,
		},
		{
			name: "should handle URL with short expiration",
			requestBody: request.ShortenURLRequest{
				Url: "https://google.com",
				Exp: 60,
			},
			setupMock: func(svc *mocks.LinkService) {
				svc.On("ShortenURL", request.ShortenURLRequest{
					Url: "https://google.com",
					Exp: 60,
				}).Return("xyz7890", nil).Once()
			},
			expectedCode:    http.StatusOK,
			expectedMessage: "Shorten URL generated successfully",
			expectedCodeLen: 7,
			expectError:     false,
		},
		{
			name: "should handle URL with long expiration",
			requestBody: request.ShortenURLRequest{
				Url: "https://github.com/user/repo",
				Exp: 86400,
			},
			setupMock: func(svc *mocks.LinkService) {
				svc.On("ShortenURL", request.ShortenURLRequest{
					Url: "https://github.com/user/repo",
					Exp: 86400,
				}).Return("long99x", nil).Once()
			},
			expectedCode:    http.StatusOK,
			expectedMessage: "Shorten URL generated successfully",
			expectedCodeLen: 7,
			expectError:     false,
		},
		{
			name: "should handle URL with zero expiration",
			requestBody: request.ShortenURLRequest{
				Url: "https://example.com/no-exp",
				Exp: 0,
			},
			setupMock: func(svc *mocks.LinkService) {
				svc.On("ShortenURL", request.ShortenURLRequest{
					Url: "https://example.com/no-exp",
					Exp: 0,
				}).Return("noexp01", nil).Once()
			},
			expectedCode:    http.StatusOK,
			expectedMessage: "Shorten URL generated successfully",
			expectedCodeLen: 7,
			expectError:     false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gin.SetMode(gin.TestMode)

			mockService := mocks.NewLinkService(t)
			tc.setupMock(mockService)

			handler := NewLinkHandler(mockService)

			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)

			// Prepare request body
			body, err := json.Marshal(tc.requestBody)
			require.NoError(t, err)

			httpRequest := httptest.NewRequest(
				http.MethodPost,
				"/links/shorten-url",
				strings.NewReader(string(body)),
			)
			httpRequest.Header.Set("Content-Type", "application/json")

			ctx.Request = httpRequest

			handler.ShortenURL(ctx)

			assert.Equal(t, tc.expectedCode, recorder.Code)

			if tc.expectError {
				var errResponse map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &errResponse)
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedMessage, errResponse["error"])
			} else {
				var successResponse response.ShortenURLResponse
				err := json.Unmarshal(recorder.Body.Bytes(), &successResponse)
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedMessage, successResponse.Message)
				assert.Equal(t, tc.expectedCodeLen, len(successResponse.Code))
				assert.NotEmpty(t, successResponse.Code)
			}
		})
	}
}

func TestLinkHandler_ShortenURL_InvalidRequest(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		requestBody   string
		expectedCode  int
		expectedError string
		setupMock     func(*mocks.LinkService)
	}{
		{
			name:          "should return 400 for invalid JSON",
			requestBody:   `{invalid json}`,
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid request body",
			setupMock:     func(svc *mocks.LinkService) {},
		},
		{
			name:          "should return 400 for malformed JSON array",
			requestBody:   `["not", "a", "valid", "request"`,
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid request body",
			setupMock:     func(svc *mocks.LinkService) {},
		},
		{
			name:          "should return 400 for empty request body",
			requestBody:   ``,
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid request body",
			setupMock:     func(svc *mocks.LinkService) {},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gin.SetMode(gin.TestMode)

			mockService := mocks.NewLinkService(t)
			tc.setupMock(mockService)

			handler := NewLinkHandler(mockService)

			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)

			httpRequest := httptest.NewRequest(
				http.MethodPost,
				"/links/shorten-url",
				strings.NewReader(tc.requestBody),
			)
			httpRequest.Header.Set("Content-Type", "application/json")

			ctx.Request = httpRequest

			handler.ShortenURL(ctx)

			assert.Equal(t, tc.expectedCode, recorder.Code)

			var errResponse map[string]interface{}
			err := json.Unmarshal(recorder.Body.Bytes(), &errResponse)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedError, errResponse["error"])
		})
	}
}

func TestLinkHandler_ShortenURL_ValidationFailure(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		requestBody   request.ShortenURLRequest
		expectedCode  int
		expectedError string
		setupMock     func(*mocks.LinkService)
	}{
		{
			name: "should return 400 for missing URL",
			requestBody: request.ShortenURLRequest{
				Url: "",
				Exp: 3600,
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Validation failed:",
			setupMock:     func(svc *mocks.LinkService) {},
		},
		{
			name: "should return 400 for invalid URL format",
			requestBody: request.ShortenURLRequest{
				Url: "not-a-valid-url",
				Exp: 3600,
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Validation failed:",
			setupMock:     func(svc *mocks.LinkService) {},
		},
		{
			name: "should return 400 for negative expiration",
			requestBody: request.ShortenURLRequest{
				Url: "https://example.com",
				Exp: -1,
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Validation failed:",
			setupMock:     func(svc *mocks.LinkService) {},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gin.SetMode(gin.TestMode)

			mockService := mocks.NewLinkService(t)
			tc.setupMock(mockService)

			handler := NewLinkHandler(mockService)

			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)

			// Prepare request body
			body, err := json.Marshal(tc.requestBody)
			require.NoError(t, err)

			httpRequest := httptest.NewRequest(
				http.MethodPost,
				"/links/shorten",
				strings.NewReader(string(body)),
			)
			httpRequest.Header.Set("Content-Type", "application/json")

			ctx.Request = httpRequest

			handler.ShortenURL(ctx)

			assert.Equal(t, tc.expectedCode, recorder.Code)

			var errResponse map[string]interface{}
			err = json.Unmarshal(recorder.Body.Bytes(), &errResponse)
			assert.NoError(t, err)
			assert.Contains(t, errResponse["error"].(string), tc.expectedError)
		})
	}
}

func TestLinkHandler_ShortenURL_SpecialCharactersInURL(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		requestBody request.ShortenURLRequest
		setupMock   func(*mocks.LinkService)
	}{
		{
			name: "should handle URL with query parameters",
			requestBody: request.ShortenURLRequest{
				Url: "https://example.com?key=value&foo=bar",
				Exp: 3600,
			},
			setupMock: func(svc *mocks.LinkService) {
				svc.On("ShortenURL", request.ShortenURLRequest{
					Url: "https://example.com?key=value&foo=bar",
					Exp: 3600,
				}).Return("query01", nil).Once()
			},
		},
		{
			name: "should handle URL with hash fragment",
			requestBody: request.ShortenURLRequest{
				Url: "https://example.com#section",
				Exp: 3600,
			},
			setupMock: func(svc *mocks.LinkService) {
				svc.On("ShortenURL", request.ShortenURLRequest{
					Url: "https://example.com#section",
					Exp: 3600,
				}).Return("hash001", nil).Once()
			},
		},
		{
			name: "should handle very long URL",
			requestBody: request.ShortenURLRequest{
				Url: "https://example.com/very/long/path/with/many/segments?param1=value1&param2=value2&param3=value3",
				Exp: 3600,
			},
			setupMock: func(svc *mocks.LinkService) {
				svc.On("ShortenURL", request.ShortenURLRequest{
					Url: "https://example.com/very/long/path/with/many/segments?param1=value1&param2=value2&param3=value3",
					Exp: 3600,
				}).Return("verylong", nil).Once()
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gin.SetMode(gin.TestMode)

			mockService := mocks.NewLinkService(t)
			tc.setupMock(mockService)

			handler := NewLinkHandler(mockService)

			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)

			body, err := json.Marshal(tc.requestBody)
			require.NoError(t, err)

			httpRequest := httptest.NewRequest(
				http.MethodPost,
				"/links/shorten",
				strings.NewReader(string(body)),
			)
			httpRequest.Header.Set("Content-Type", "application/json")

			ctx.Request = httpRequest

			handler.ShortenURL(ctx)

			assert.Equal(t, http.StatusOK, recorder.Code)

			var successResponse response.ShortenURLResponse
			err = json.Unmarshal(recorder.Body.Bytes(), &successResponse)
			assert.NoError(t, err)
			assert.NotEmpty(t, successResponse.Code)
			assert.Equal(t, "Shorten URL generated successfully", successResponse.Message)
		})
	}
}

func TestLinkHandler_NewLinkHandler(t *testing.T) {
	t.Parallel()

	mockService := mocks.NewLinkService(t)

	handler := NewLinkHandler(mockService)

	assert.NotNil(t, handler)
	assert.Implements(t, (*Link)(nil), handler)
}

func TestLinkHandler_ShortenURL_ResponseStructure(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	mockService := mocks.NewLinkService(t)
	mockService.On("ShortenURL", request.ShortenURLRequest{
		Url: "https://example.com",
		Exp: 3600,
	}).Return("testcode", nil).Once()

	handler := NewLinkHandler(mockService)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)

	reqBody := request.ShortenURLRequest{
		Url: "https://example.com",
		Exp: 3600,
	}
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	httpRequest := httptest.NewRequest(
		http.MethodPost,
		"/links/shorten-url",
		strings.NewReader(string(body)),
	)
	httpRequest.Header.Set("Content-Type", "application/json")

	ctx.Request = httpRequest

	handler.ShortenURL(ctx)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var successResponse response.ShortenURLResponse
	err = json.Unmarshal(recorder.Body.Bytes(), &successResponse)
	assert.NoError(t, err)

	// Verify response has both required fields
	assert.Equal(t, "testcode", successResponse.Code)
	assert.Equal(t, "Shorten URL generated successfully", successResponse.Message)
}

func TestLinkHandler_RedirectToURL(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                string
		code                string
		setupMock           func(*mocks.LinkService)
		expectedCode        int
		expectedLocationURL string
		expectError         bool
	}{
		{
			name: "should redirect successfully when URL is found",
			code: "abc123d",
			setupMock: func(svc *mocks.LinkService) {
				svc.On("GetOriginalURL", "abc123d").Return("https://example.com/very/long/path", nil).Once()
			},
			expectedCode:        http.StatusFound,
			expectedLocationURL: "https://example.com/very/long/path",
			expectError:         false,
		},
		{
			name: "should return 404 when code is not found",
			code: "notfound",
			setupMock: func(svc *mocks.LinkService) {
				svc.On("GetOriginalURL", "notfound").Return("", errors.New("link not found")).Once()
			},
			expectedCode:        http.StatusNotFound,
			expectedLocationURL: "",
			expectError:         true,
		},
		{
			name: "should redirect to different URL",
			code: "xyz7890",
			setupMock: func(svc *mocks.LinkService) {
				svc.On("GetOriginalURL", "xyz7890").Return("https://github.com/huypham67/bookmark-management", nil).Once()
			},
			expectedCode:        http.StatusFound,
			expectedLocationURL: "https://github.com/huypham67/bookmark-management",
			expectError:         false,
		},
		{
			name: "should handle URL with query parameters",
			code: "query01",
			setupMock: func(svc *mocks.LinkService) {
				svc.On("GetOriginalURL", "query01").Return("https://example.com?key=value&foo=bar", nil).Once()
			},
			expectedCode:        http.StatusFound,
			expectedLocationURL: "https://example.com?key=value&foo=bar",
			expectError:         false,
		},
		{
			name: "should handle URL with hash fragment",
			code: "hash001",
			setupMock: func(svc *mocks.LinkService) {
				svc.On("GetOriginalURL", "hash001").Return("https://example.com#section", nil).Once()
			},
			expectedCode:        http.StatusFound,
			expectedLocationURL: "https://example.com#section",
			expectError:         false,
		},
		{
			name: "should handle very long URL",
			code: "verylong",
			setupMock: func(svc *mocks.LinkService) {
				svc.On("GetOriginalURL", "verylong").Return("https://example.com/very/long/path/with/many/segments?param1=value1&param2=value2&param3=value3", nil).Once()
			},
			expectedCode:        http.StatusFound,
			expectedLocationURL: "https://example.com/very/long/path/with/many/segments?param1=value1&param2=value2&param3=value3",
			expectError:         false,
		},
		{
			name: "should handle redis connection error",
			code: "error01",
			setupMock: func(svc *mocks.LinkService) {
				svc.On("GetOriginalURL", "error01").Return("", errors.New("redis connection error")).Once()
			},
			expectedCode:        http.StatusNotFound,
			expectedLocationURL: "",
			expectError:         true,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gin.SetMode(gin.TestMode)

			mockService := mocks.NewLinkService(t)
			tc.setupMock(mockService)

			handler := NewLinkHandler(mockService)

			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)

			httpRequest := httptest.NewRequest(
				http.MethodGet,
				"/links/redirect/"+tc.code,
				nil,
			)

			ctx.Request = httpRequest
			ctx.Params = append(ctx.Params, gin.Param{Key: "code", Value: tc.code})

			handler.RedirectToURL(ctx)

			assert.Equal(t, tc.expectedCode, recorder.Code)

			if tc.expectError {
				var errResponse map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &errResponse)
				assert.NoError(t, err)
				assert.Equal(t, "Short link not found", errResponse["error"])
			} else {
				assert.Equal(t, tc.expectedLocationURL, recorder.Header().Get("Location"))
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestLinkHandler_RedirectToURL_EmptyCode(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	mockService := mocks.NewLinkService(t)
	mockService.On("GetOriginalURL", "").Return("", errors.New("empty code")).Once()

	handler := NewLinkHandler(mockService)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)

	httpRequest := httptest.NewRequest(
		http.MethodGet,
		"/links/redirect/",
		nil,
	)

	ctx.Request = httpRequest
	ctx.Params = append(ctx.Params, gin.Param{Key: "code", Value: ""})

	handler.RedirectToURL(ctx)

	assert.Equal(t, http.StatusNotFound, recorder.Code)

	var errResponse map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &errResponse)
	assert.NoError(t, err)
	assert.Equal(t, "Short link not found", errResponse["error"])
}
